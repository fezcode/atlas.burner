package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fezcode/atlas.burner/internal/catalog"
	"github.com/fezcode/atlas.burner/internal/downloader"
	"github.com/fezcode/atlas.burner/internal/usb"
)

// ── Steps ───────────────────────────────────────────────────────
type step int

const (
	stepUSB step = iota
	stepFamily
	stepOS
	stepArch      // pick architecture
	stepSource    // pick image source: download or local file
	stepDestDir   // pick download destination directory
	stepDownload  // download progress
	stepOptions   // burn options
	stepConfirm   // confirm before burn
	stepBurning   // burning progress
	stepDone      // finished
)

var stepNames = []string{
	"USB Device",
	"OS Family",
	"Operating System",
	"Architecture",
	"Image Source",
	"Download Path",
	"Download",
	"Burn Options",
	"Confirm",
	"Burn",
	"Done",
}

// ── Burn Options ────────────────────────────────────────────────
type BurnOptions struct {
	BlockSize     int    // bytes
	Verify        bool   // verify checksum after burn
	ForceUnmount  bool   // unmount device before burn
	PartitionType string // "keep" | "mbr" | "gpt"
	Label         string // volume label
}

var blockSizeOptions = []struct {
	label string
	value int
}{
	{"512 B", 512},
	{"4 KB", 4 * 1024},
	{"64 KB", 64 * 1024},
	{"1 MB", 1024 * 1024},
	{"4 MB (default)", 4 * 1024 * 1024},
	{"8 MB", 8 * 1024 * 1024},
}

var partitionOptions = []string{"Keep original", "MBR", "GPT"}

// ── Messages ────────────────────────────────────────────────────
type usbLoadedMsg struct {
	devices []usb.Device
	err     error
}

type downloadTickMsg time.Time

type downloadDoneMsg struct{ err error }

type burnTickMsg time.Time

type burnDoneMsg struct{ err error }

// ── Model ───────────────────────────────────────────────────────
type Model struct {
	step    step
	width   int
	height  int
	err     error
	quitted bool

	// USB
	usbDevices  []usb.Device
	usbCursor   int
	usbLoading  bool
	selectedUSB *usb.Device

	// OS Family
	familyCursor   int
	selectedFamily catalog.Family

	// OS
	osCursor   int
	filteredOS []catalog.OS
	selectedOS *catalog.OS

	// Architecture
	archCursor      int
	availableArchs  []catalog.Arch
	selectedVariant *catalog.Variant

	// Source
	sourceCursor int // 0 = download, 1 = local file
	localPath    string
	editingPath  bool

	// Download destination
	destDir     string
	editingDest bool

	// Download
	dl           *downloader.Downloader
	dlProgress   float64
	downloading  bool
	downloadDone bool

	// Burn options
	options      BurnOptions
	optCursor    int // which option is focused
	optSubCursor int // cursor within a sub-option (e.g., block size list)
	optEditing   bool

	// Burn
	burnProgress *float64 // shared pointer so the goroutine callback can update it
	burnStart    time.Time
	burning      bool
	burnDone     bool
}

func NewModel() Model {
	home, _ := os.UserHomeDir()
	return Model{
		step:       stepUSB,
		usbLoading: true,
		options: BurnOptions{
			BlockSize:     4 * 1024 * 1024,
			Verify:        true,
			ForceUnmount:  true,
			PartitionType: "keep",
		},
		destDir:      filepath.Join(home, "Downloads"),
		burnProgress: new(float64),
		familyCursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return loadUSBDevices
}

func loadUSBDevices() tea.Msg {
	devices, err := usb.GetRemovableDevices()
	return usbLoadedMsg{devices: devices, err: err}
}

func tickDownload() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return downloadTickMsg(t)
	})
}

func tickBurn() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return burnTickMsg(t)
	})
}

// ── Update ──────────────────────────────────────────────────────
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			m.quitted = true
			return m, tea.Quit
		}
		return m.handleKey(msg)

	case usbLoadedMsg:
		m.usbLoading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.usbDevices = msg.devices
		}
		return m, nil

	case downloadTickMsg:
		if m.dl != nil {
			m.dlProgress = m.dl.Progress()
			if m.dl.IsComplete() {
				m.downloading = false
				m.downloadDone = true
				if err := m.dl.Err(); err != nil {
					m.err = err
				}
				return m, nil
			}
		}
		return m, tickDownload()

	case downloadDoneMsg:
		m.downloading = false
		m.downloadDone = true
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil

	case burnTickMsg:
		if m.burnDone {
			return m, nil
		}
		return m, tickBurn()

	case burnDoneMsg:
		m.burning = false
		m.burnDone = true
		if msg.err != nil {
			m.err = msg.err
		}
		m.step = stepDone
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Back navigation (except when editing text)
	if key == "esc" && !m.editingPath && !m.editingDest && !m.optEditing {
		if m.step > stepUSB && m.step < stepBurning {
			m.step--
			// Skip download step when going back if not downloading
			if m.step == stepDownload && !m.downloading {
				m.step = stepDestDir
			}
			m.err = nil
			return m, nil
		}
	}

	switch m.step {
	case stepUSB:
		return m.handleUSBKey(key)
	case stepFamily:
		return m.handleFamilyKey(key)
	case stepOS:
		return m.handleOSKey(key)
	case stepArch:
		return m.handleArchKey(key)
	case stepSource:
		return m.handleSourceKey(key, msg)
	case stepDestDir:
		return m.handleDestDirKey(key, msg)
	case stepDownload:
		if m.downloadDone && m.err == nil {
			m.step = stepOptions
		}
		return m, nil
	case stepOptions:
		return m.handleOptionsKey(key, msg)
	case stepConfirm:
		return m.handleConfirmKey(key)
	case stepBurning:
		return m, nil
	case stepDone:
		if key == "q" || key == "enter" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) handleUSBKey(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.usbCursor > 0 {
			m.usbCursor--
		}
	case "down", "j":
		if m.usbCursor < len(m.usbDevices)-1 {
			m.usbCursor++
		}
	case "r":
		m.usbLoading = true
		m.usbDevices = nil
		return m, loadUSBDevices
	case "enter":
		if len(m.usbDevices) > 0 {
			d := m.usbDevices[m.usbCursor]
			m.selectedUSB = &d
			m.step = stepFamily
		}
	}
	return m, nil
}

func (m Model) handleFamilyKey(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.familyCursor > 0 {
			m.familyCursor--
		}
	case "down", "j":
		if m.familyCursor < len(catalog.Families)-1 {
			m.familyCursor++
		}
	case "enter":
		m.selectedFamily = catalog.Families[m.familyCursor]
		m.filteredOS = catalog.OSByFamily(m.selectedFamily)
		m.osCursor = 0
		m.step = stepOS
	}
	return m, nil
}

func (m Model) handleOSKey(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.osCursor > 0 {
			m.osCursor--
		}
	case "down", "j":
		if m.osCursor < len(m.filteredOS)-1 {
			m.osCursor++
		}
	case "enter":
		if len(m.filteredOS) > 0 {
			selected := m.filteredOS[m.osCursor]
			m.selectedOS = &selected
			m.availableArchs = selected.Archs()
			m.archCursor = 0
			if len(m.availableArchs) == 1 {
				// Only one architecture — auto-select and skip
				v := selected.VariantByArch(m.availableArchs[0])
				m.selectedVariant = v
				m.step = stepSource
			} else {
				m.step = stepArch
			}
		}
	}
	return m, nil
}

func (m Model) handleArchKey(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.archCursor > 0 {
			m.archCursor--
		}
	case "down", "j":
		if m.archCursor < len(m.availableArchs)-1 {
			m.archCursor++
		}
	case "enter":
		if len(m.availableArchs) > 0 {
			arch := m.availableArchs[m.archCursor]
			v := m.selectedOS.VariantByArch(arch)
			m.selectedVariant = v
			m.step = stepSource
		}
	}
	return m, nil
}

func (m Model) handleSourceKey(key string, msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.editingPath {
		switch key {
		case "enter":
			m.editingPath = false
			if m.localPath != "" {
				if _, err := os.Stat(m.localPath); err != nil {
					m.err = fmt.Errorf("file not found: %s", m.localPath)
					return m, nil
				}
				m.err = nil
				m.step = stepOptions
			}
		case "esc":
			m.editingPath = false
			m.localPath = ""
		case "backspace":
			if len(m.localPath) > 0 {
				m.localPath = m.localPath[:len(m.localPath)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				m.localPath += string(msg.Runes)
			}
		}
		return m, nil
	}

	switch key {
	case "up", "k":
		if m.sourceCursor > 0 {
			m.sourceCursor--
		}
	case "down", "j":
		if m.sourceCursor < 1 {
			m.sourceCursor++
		}
	case "enter":
		if m.sourceCursor == 0 {
			m.step = stepDestDir
		} else {
			m.editingPath = true
			m.localPath = ""
		}
	}
	return m, nil
}

func (m Model) handleDestDirKey(key string, msg tea.KeyMsg) (Model, tea.Cmd) {
	if !m.editingDest {
		switch key {
		case "enter":
			return m.startDownload()
		case "e":
			m.editingDest = true
		}
		return m, nil
	}

	switch key {
	case "enter":
		m.editingDest = false
	case "esc":
		m.editingDest = false
	case "backspace":
		if len(m.destDir) > 0 {
			m.destDir = m.destDir[:len(m.destDir)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.destDir += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) startDownload() (Model, tea.Cmd) {
	if m.selectedVariant == nil {
		return m, nil
	}

	dest := filepath.Join(m.destDir, sanitizeFilename(m.selectedOS.Name+"-"+m.selectedOS.Version)+".iso")
	m.dl = downloader.New()
	m.downloading = true
	m.downloadDone = false
	m.dlProgress = 0
	m.step = stepDownload
	m.localPath = dest

	url := m.selectedVariant.DownloadURL

	return m, tea.Batch(
		func() tea.Msg {
			err := m.dl.Start(url, dest)
			if err != nil {
				return downloadDoneMsg{err: err}
			}
			err = m.dl.Wait()
			return downloadDoneMsg{err: err}
		},
		tickDownload(),
	)
}

func (m Model) handleOptionsKey(key string, msg tea.KeyMsg) (Model, tea.Cmd) {
	optionCount := 5

	if m.optEditing {
		switch m.optCursor {
		case 0: // block size
			switch key {
			case "up", "k":
				if m.optSubCursor > 0 {
					m.optSubCursor--
				}
			case "down", "j":
				if m.optSubCursor < len(blockSizeOptions)-1 {
					m.optSubCursor++
				}
			case "enter":
				m.options.BlockSize = blockSizeOptions[m.optSubCursor].value
				m.optEditing = false
			case "esc":
				m.optEditing = false
			}
		case 3: // partition type
			switch key {
			case "up", "k":
				if m.optSubCursor > 0 {
					m.optSubCursor--
				}
			case "down", "j":
				if m.optSubCursor < len(partitionOptions)-1 {
					m.optSubCursor++
				}
			case "enter":
				val := []string{"keep", "mbr", "gpt"}
				m.options.PartitionType = val[m.optSubCursor]
				m.optEditing = false
			case "esc":
				m.optEditing = false
			}
		case 4: // label
			switch key {
			case "enter":
				m.optEditing = false
			case "esc":
				m.optEditing = false
			case "backspace":
				if len(m.options.Label) > 0 {
					m.options.Label = m.options.Label[:len(m.options.Label)-1]
				}
			default:
				if len(msg.Runes) > 0 {
					m.options.Label += string(msg.Runes)
				}
			}
		}
		return m, nil
	}

	switch key {
	case "up", "k":
		if m.optCursor > 0 {
			m.optCursor--
		}
	case "down", "j":
		if m.optCursor < optionCount-1 {
			m.optCursor++
		}
	case "enter":
		switch m.optCursor {
		case 0:
			m.optEditing = true
			for i, bs := range blockSizeOptions {
				if bs.value == m.options.BlockSize {
					m.optSubCursor = i
					break
				}
			}
		case 1:
			m.options.Verify = !m.options.Verify
		case 2:
			m.options.ForceUnmount = !m.options.ForceUnmount
		case 3:
			m.optEditing = true
			switch m.options.PartitionType {
			case "keep":
				m.optSubCursor = 0
			case "mbr":
				m.optSubCursor = 1
			case "gpt":
				m.optSubCursor = 2
			}
		case 4:
			m.optEditing = true
		}
	case " ":
		switch m.optCursor {
		case 1:
			m.options.Verify = !m.options.Verify
		case 2:
			m.options.ForceUnmount = !m.options.ForceUnmount
		}
	case "tab", "n":
		m.step = stepConfirm
	}
	return m, nil
}

func (m Model) handleConfirmKey(key string) (Model, tea.Cmd) {
	switch key {
	case "y", "Y":
		return m.startBurning()
	case "n", "N", "esc":
		m.step = stepOptions
	}
	return m, nil
}

func (m Model) startBurning() (Model, tea.Cmd) {
	if m.selectedUSB == nil || m.localPath == "" {
		return m, nil
	}

	m.burning = true
	m.burnDone = false
	*m.burnProgress = 0
	m.burnStart = time.Now()
	m.step = stepBurning

	devicePath := m.selectedUSB.Path
	imagePath := m.localPath
	progressPtr := m.burnProgress // shared pointer between goroutine and model

	return m, tea.Batch(
		func() tea.Msg {
			err := usb.BurnImage(imagePath, devicePath, func(progress float64) {
				*progressPtr = progress
			})
			return burnDoneMsg{err: err}
		},
		tickBurn(),
	)
}

// ── View ────────────────────────────────────────────────────────
func (m Model) View() string {
	if m.quitted {
		return mutedStyle.Render("  Cancelled.\n")
	}

	var s strings.Builder

	s.WriteString(banner())
	s.WriteString("\n")
	s.WriteString(renderStepBar(m.step))
	s.WriteString("\n\n")

	switch m.step {
	case stepUSB:
		s.WriteString(m.viewUSB())
	case stepFamily:
		s.WriteString(m.viewFamily())
	case stepOS:
		s.WriteString(m.viewOS())
	case stepArch:
		s.WriteString(m.viewArch())
	case stepSource:
		s.WriteString(m.viewSource())
	case stepDestDir:
		s.WriteString(m.viewDestDir())
	case stepDownload:
		s.WriteString(m.viewDownload())
	case stepOptions:
		s.WriteString(m.viewOptions())
	case stepConfirm:
		s.WriteString(m.viewConfirm())
	case stepBurning:
		s.WriteString(m.viewBurning())
	case stepDone:
		s.WriteString(m.viewDone())
	}

	if m.err != nil && m.step != stepDone {
		s.WriteString("\n")
		s.WriteString(dangerStyle.Render("  Error: " + m.err.Error()))
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render(m.helpText()))

	return appStyle.Render(s.String())
}

func (m Model) helpText() string {
	switch m.step {
	case stepUSB:
		return "  ↑/↓ navigate  •  enter select  •  r refresh  •  ctrl+c quit"
	case stepFamily, stepOS, stepArch:
		return "  ↑/↓ navigate  •  enter select  •  esc back  •  ctrl+c quit"
	case stepSource:
		if m.editingPath {
			return "  type path  •  enter confirm  •  esc cancel"
		}
		return "  ↑/↓ navigate  •  enter select  •  esc back  •  ctrl+c quit"
	case stepDestDir:
		if m.editingDest {
			return "  type path  •  enter confirm  •  esc cancel"
		}
		return "  enter download  •  e edit path  •  esc back  •  ctrl+c quit"
	case stepDownload:
		return "  downloading...  •  ctrl+c abort"
	case stepOptions:
		if m.optEditing {
			return "  ↑/↓ navigate  •  enter select  •  esc cancel"
		}
		return "  ↑/↓ navigate  •  enter/space toggle  •  tab/n next  •  esc back  •  ctrl+c quit"
	case stepConfirm:
		return "  y confirm  •  n/esc go back  •  ctrl+c quit"
	case stepBurning:
		return "  burning in progress...  •  ctrl+c abort (not recommended)"
	case stepDone:
		return "  enter/q exit"
	}
	return "  ctrl+c quit"
}

// ── Step bar ────────────────────────────────────────────────────
func renderStepBar(current step) string {
	var parts []string
	visible := []step{stepUSB, stepFamily, stepOS, stepArch, stepSource, stepOptions, stepConfirm, stepBurning, stepDone}
	for _, s := range visible {
		name := stepNames[s]
		if s < current {
			parts = append(parts, completedStepStyle.Render("✓ "+name))
		} else if s == current {
			parts = append(parts, activeStepStyle.Render("● "+name))
		} else {
			parts = append(parts, inactiveStepStyle.Render("○ "+name))
		}
	}
	return stepIndicatorStyle.Render("  " + strings.Join(parts, mutedStyle.Render("  ›  ")))
}

// ── USB View ────────────────────────────────────────────────────
func (m Model) viewUSB() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Select USB Device"))
	s.WriteString("\n\n")

	if m.usbLoading {
		s.WriteString(mutedStyle.Render("  Scanning for USB devices..."))
		return s.String()
	}

	if len(m.usbDevices) == 0 {
		content := mutedStyle.Render("No removable USB devices found.") + "\n\n" +
			keyStyle.Render("r") + mutedStyle.Render(" to refresh")
		s.WriteString(panelStyle.Render(content))
		return s.String()
	}

	for i, dev := range m.usbDevices {
		sizeGB := float64(dev.Size) / (1024 * 1024 * 1024)
		bootTag := ""
		switch dev.Bootable {
		case usb.BootMBR:
			bootTag = successStyle.Render("  [Bootable: MBR]")
		case usb.BootGPT:
			bootTag = successStyle.Render("  [Bootable: GPT/UEFI]")
		case usb.BootUnknown:
			bootTag = mutedStyle.Render("  [Bootable: ?]")
		}
		name := fmt.Sprintf("%s  (%s)", dev.Name, dev.Description)
		detail := fmt.Sprintf("%.1f GB  •  %s", sizeGB, dev.Path)

		if i == m.usbCursor {
			s.WriteString(selectedItemStyle.Render("▸ " + name))
			s.WriteString(bootTag)
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		} else {
			s.WriteString(normalItemStyle.Render("  " + name))
			s.WriteString(bootTag)
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		}
		s.WriteString("\n")
	}
	return s.String()
}

// ── Family View ─────────────────────────────────────────────────
func (m Model) viewFamily() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Select OS Family"))
	s.WriteString("\n\n")

	icons := map[catalog.Family]string{
		catalog.FamilyLinux:   "🐧",
		catalog.FamilyWindows: "🪟",
		catalog.FamilyUnix:    "👹",
	}

	for i, fam := range catalog.Families {
		icon := icons[fam]
		label := fmt.Sprintf("%s  %s", icon, string(fam))
		count := len(catalog.OSByFamily(fam))
		detail := fmt.Sprintf("%d distributions available", count)

		if i == m.familyCursor {
			s.WriteString(selectedItemStyle.Render("▸ " + label))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		} else {
			s.WriteString(normalItemStyle.Render("  " + label))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		}
		s.WriteString("\n")
	}
	return s.String()
}

// ── OS View ─────────────────────────────────────────────────────
func (m Model) viewOS() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("Select %s Distribution", m.selectedFamily)))
	s.WriteString("\n\n")

	if len(m.filteredOS) == 0 {
		s.WriteString(mutedStyle.Render("  No distributions found."))
		return s.String()
	}

	for i, o := range m.filteredOS {
		name := fmt.Sprintf("%s %s", o.Name, o.Version)
		archCount := len(o.Variants)
		archLabel := fmt.Sprintf("%d arch", archCount)
		if archCount > 1 {
			archLabel += "s"
		}
		detail := fmt.Sprintf("%s  •  %s  •  %s", o.Variants[0].Size, o.Format, archLabel)

		if i == m.osCursor {
			s.WriteString(selectedItemStyle.Render("▸ " + name))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		} else {
			s.WriteString(normalItemStyle.Render("  " + name))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		}
		s.WriteString("\n")
	}
	return s.String()
}

// ── Arch View ───────────────────────────────────────────────────
func (m Model) viewArch() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Select Architecture"))
	s.WriteString("\n\n")

	if m.selectedOS != nil {
		s.WriteString(subtitleStyle.Render("  "+m.selectedOS.Name+" "+m.selectedOS.Version))
		s.WriteString("\n\n")
	}

	for i, arch := range m.availableArchs {
		v := m.selectedOS.VariantByArch(arch)
		label := string(arch)
		detail := v.Size + "  •  " + v.DownloadURL

		if i == m.archCursor {
			s.WriteString(selectedItemStyle.Render("▸ " + label))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		} else {
			s.WriteString(normalItemStyle.Render("  " + label))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + detail))
		}
		s.WriteString("\n")
	}
	return s.String()
}

// ── Source View ──────────────────────────────────────────────────
func (m Model) viewSource() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Image Source"))
	s.WriteString("\n\n")

	dlURL := ""
	if m.selectedVariant != nil {
		dlURL = m.selectedVariant.DownloadURL
	}

	options := []struct {
		label string
		desc  string
	}{
		{"Download from Internet", dlURL},
		{"Use local ISO/IMG file", "Browse for an existing image file"},
	}

	for i, opt := range options {
		if i == m.sourceCursor {
			s.WriteString(selectedItemStyle.Render("▸ " + opt.label))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + opt.desc))
		} else {
			s.WriteString(normalItemStyle.Render("  " + opt.label))
			s.WriteString("\n")
			s.WriteString(itemDescStyle.Render("  " + opt.desc))
		}
		s.WriteString("\n")
	}

	if m.editingPath {
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("  Path: "))
		s.WriteString(valueStyle.Render(m.localPath + "█"))
	}

	return s.String()
}

// ── Dest Dir View ───────────────────────────────────────────────
func (m Model) viewDestDir() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Download Destination"))
	s.WriteString("\n\n")

	content := labelStyle.Render("Directory: ") + "\n"
	if m.editingDest {
		content += valueStyle.Render(m.destDir+"█") + "\n"
	} else {
		content += valueStyle.Render(m.destDir) + "\n"
	}
	content += "\n"
	if m.selectedOS != nil {
		filename := sanitizeFilename(m.selectedOS.Name+"-"+m.selectedOS.Version) + ".iso"
		content += labelStyle.Render("Filename: ") + mutedStyle.Render(filename) + "\n"
	}
	if m.selectedVariant != nil {
		content += labelStyle.Render("Size:     ") + mutedStyle.Render(m.selectedVariant.Size) + "\n"
		content += labelStyle.Render("Arch:     ") + mutedStyle.Render(string(m.selectedVariant.Arch))
	}

	s.WriteString(panelStyle.Render(content))
	return s.String()
}

// ── Download View ───────────────────────────────────────────────
func (m Model) viewDownload() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Downloading"))
	s.WriteString("\n\n")

	if m.selectedOS != nil {
		label := m.selectedOS.Name + " " + m.selectedOS.Version
		if m.selectedVariant != nil {
			label += "  [" + string(m.selectedVariant.Arch) + "]"
		}
		s.WriteString(labelStyle.Render("  " + label))
		s.WriteString("\n\n")
	}

	s.WriteString(renderProgressBar(m.dlProgress, 40))
	fmt.Fprintf(&s, "  %.1f%%", m.dlProgress*100)
	s.WriteString("\n\n")

	if m.downloadDone {
		if m.err == nil {
			s.WriteString(successStyle.Render("  ✓ Download complete!"))
			s.WriteString("\n")
			s.WriteString(mutedStyle.Render("  Press any key to continue..."))
		}
	} else {
		s.WriteString(mutedStyle.Render("  Saving to: " + m.localPath))
	}

	return s.String()
}

// ── Options View ────────────────────────────────────────────────
func (m Model) viewOptions() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Burn Options"))
	s.WriteString("\n\n")

	type optionLine struct {
		label string
		value string
	}

	bsLabel := "4 MB (default)"
	for _, bs := range blockSizeOptions {
		if bs.value == m.options.BlockSize {
			bsLabel = bs.label
			break
		}
	}

	ptLabel := "Keep original"
	switch m.options.PartitionType {
	case "mbr":
		ptLabel = "MBR"
	case "gpt":
		ptLabel = "GPT"
	}

	labelVal := m.options.Label
	if labelVal == "" {
		labelVal = "(none)"
	}

	opts := []optionLine{
		{"Block Size", bsLabel},
		{"Verify After Burn", boolIcon(m.options.Verify)},
		{"Force Unmount", boolIcon(m.options.ForceUnmount)},
		{"Partition Table", ptLabel},
		{"Volume Label", labelVal},
	}

	for i, opt := range opts {
		prefix := "  "
		style := normalItemStyle
		if i == m.optCursor {
			prefix = "▸ "
			style = selectedItemStyle
		}
		line := fmt.Sprintf("%s%-20s %s", prefix, opt.label, valueStyle.Render(opt.value))
		s.WriteString(style.Render(line))
		s.WriteString("\n")

		if m.optEditing && i == m.optCursor {
			switch i {
			case 0:
				for j, bs := range blockSizeOptions {
					if j == m.optSubCursor {
						s.WriteString(selectedItemStyle.Render("      ▸ " + bs.label))
					} else {
						s.WriteString(normalItemStyle.Render("        " + bs.label))
					}
					s.WriteString("\n")
				}
			case 3:
				for j, pt := range partitionOptions {
					if j == m.optSubCursor {
						s.WriteString(selectedItemStyle.Render("      ▸ " + pt))
					} else {
						s.WriteString(normalItemStyle.Render("        " + pt))
					}
					s.WriteString("\n")
				}
			case 4:
				s.WriteString(itemDescStyle.Render("      " + m.options.Label + "█"))
				s.WriteString("\n")
			}
		}
	}

	return s.String()
}

// ── Confirm View ────────────────────────────────────────────────
func (m Model) viewConfirm() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Confirm Burn"))
	s.WriteString("\n\n")

	var content strings.Builder
	content.WriteString(dangerStyle.Render("⚠  WARNING: This will ERASE ALL DATA on the target device!"))
	content.WriteString("\n\n")

	if m.selectedOS != nil {
		content.WriteString(labelStyle.Render("Image:   "))
		content.WriteString(valueStyle.Render(m.selectedOS.Name + " " + m.selectedOS.Version))
		content.WriteString("\n")
	}
	if m.selectedVariant != nil {
		content.WriteString(labelStyle.Render("Arch:    "))
		content.WriteString(valueStyle.Render(string(m.selectedVariant.Arch)))
		content.WriteString("\n")
	}
	content.WriteString(labelStyle.Render("Source:  "))
	content.WriteString(valueStyle.Render(m.localPath))
	content.WriteString("\n")
	if m.selectedUSB != nil {
		content.WriteString(labelStyle.Render("Target:  "))
		content.WriteString(dangerStyle.Render(m.selectedUSB.String()))
		content.WriteString("\n")
	}

	bsLabel := "4 MB"
	for _, bs := range blockSizeOptions {
		if bs.value == m.options.BlockSize {
			bsLabel = bs.label
			break
		}
	}
	content.WriteString("\n")
	content.WriteString(labelStyle.Render("Block:   "))
	content.WriteString(mutedStyle.Render(bsLabel))
	content.WriteString("\n")
	content.WriteString(labelStyle.Render("Verify:  "))
	content.WriteString(mutedStyle.Render(boolIcon(m.options.Verify)))
	content.WriteString("\n")
	content.WriteString(labelStyle.Render("Unmount: "))
	content.WriteString(mutedStyle.Render(boolIcon(m.options.ForceUnmount)))
	content.WriteString("\n\n")

	content.WriteString(keyStyle.Render("y"))
	content.WriteString(mutedStyle.Render(" to confirm  •  "))
	content.WriteString(keyStyle.Render("n"))
	content.WriteString(mutedStyle.Render(" to go back"))

	s.WriteString(confirmPanelStyle.Render(content.String()))
	return s.String()
}

// ── Burning View ────────────────────────────────────────────────
func (m Model) viewBurning() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Burning Image"))
	s.WriteString("\n\n")

	if m.selectedOS != nil {
		s.WriteString(labelStyle.Render("  " + m.selectedOS.Name + " " + m.selectedOS.Version))
		if m.selectedVariant != nil {
			s.WriteString(mutedStyle.Render("  [" + string(m.selectedVariant.Arch) + "]"))
		}
		s.WriteString(mutedStyle.Render("  →  "))
		if m.selectedUSB != nil {
			s.WriteString(valueStyle.Render(m.selectedUSB.Name))
		}
		s.WriteString("\n\n")
	}

	s.WriteString(renderProgressBar(*m.burnProgress, 40))
	fmt.Fprintf(&s, "  %.1f%%", *m.burnProgress*100)
	s.WriteString("\n\n")

	elapsed := time.Since(m.burnStart).Truncate(time.Second)
	s.WriteString(mutedStyle.Render(fmt.Sprintf("  Elapsed: %s", formatDuration(elapsed))))

	// Estimate remaining time
	progress := *m.burnProgress
	if progress > 0.01 {
		eta := time.Duration(float64(elapsed) * (1 - progress) / progress).Truncate(time.Second)
		s.WriteString(mutedStyle.Render(fmt.Sprintf("  •  ETA: %s", formatDuration(eta))))
	}
	s.WriteString("\n\n")

	if m.burning {
		s.WriteString(dangerStyle.Render("  🔥 Writing to device — do NOT remove the USB drive!"))
	}

	return s.String()
}

// ── Done View ───────────────────────────────────────────────────
func (m Model) viewDone() string {
	var s strings.Builder

	if m.err != nil {
		content := dangerStyle.Render("✗ Burn failed!") + "\n\n" +
			mutedStyle.Render(m.err.Error())
		s.WriteString(panelStyle.Render(content))
	} else {
		var content strings.Builder
		content.WriteString(successStyle.Render("✓ Burn completed successfully!"))
		content.WriteString("\n\n")
		if m.selectedOS != nil {
			content.WriteString(labelStyle.Render("Image:  "))
			content.WriteString(valueStyle.Render(m.selectedOS.Name + " " + m.selectedOS.Version))
			if m.selectedVariant != nil {
				content.WriteString(mutedStyle.Render("  [" + string(m.selectedVariant.Arch) + "]"))
			}
			content.WriteString("\n")
		}
		if m.selectedUSB != nil {
			content.WriteString(labelStyle.Render("Device: "))
			content.WriteString(valueStyle.Render(m.selectedUSB.String()))
			content.WriteString("\n")
		}
		content.WriteString("\n")
		content.WriteString(mutedStyle.Render("You can now safely remove the USB device and boot from it."))
		s.WriteString(successPanelStyle.Render(content.String()))
	}

	return s.String()
}

// ── Helpers ─────────────────────────────────────────────────────
func renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := progressFullStyle.Render(strings.Repeat("█", filled)) +
		progressEmptyStyle.Render(strings.Repeat("░", empty))

	return "  " + bar
}

func boolIcon(v bool) string {
	if v {
		return successStyle.Render("✓ Yes")
	}
	return mutedStyle.Render("✗ No")
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func sanitizeFilename(name string) string {
	r := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"!", "",
		"'", "",
	)
	return strings.ToLower(r.Replace(name))
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
