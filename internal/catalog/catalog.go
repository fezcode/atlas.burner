package catalog

type Family string

const (
	FamilyWindows Family = "Windows"
	FamilyLinux   Family = "Linux"
	FamilyUnix    Family = "Unix / BSD"
)

type Arch string

const (
	ArchAMD64   Arch = "x86_64 (amd64)"
	ArchARM64   Arch = "ARM64 (aarch64)"
	ArchI386    Arch = "x86 (i386)"
	ArchRISCV64 Arch = "RISC-V 64"
)

// Variant is a specific architecture build of an OS.
type Variant struct {
	Arch        Arch
	DownloadURL string
	Size        string
}

type OS struct {
	Name     string
	Family   Family
	Version  string
	Format   string // iso, img, etc.
	Variants []Variant
}

func (o OS) FilterValue() string {
	return o.Name + " " + o.Version
}

func (o OS) Title() string {
	return o.Name + " " + o.Version
}

func (o OS) Description() string {
	if len(o.Variants) > 0 {
		return o.Variants[0].Size + " | " + o.Format
	}
	return o.Format
}

// Archs returns the list of architectures available for this OS.
func (o OS) Archs() []Arch {
	var archs []Arch
	for _, v := range o.Variants {
		archs = append(archs, v.Arch)
	}
	return archs
}

// VariantByArch returns the variant for the given architecture, or nil.
func (o OS) VariantByArch(arch Arch) *Variant {
	for _, v := range o.Variants {
		if v.Arch == arch {
			return &v
		}
	}
	return nil
}

var Families = []Family{
	FamilyLinux,
	FamilyWindows,
	FamilyUnix,
}

var Catalog = []OS{
	// ── Linux ───────────────────────────────────────────────────
	{
		Name:    "Ubuntu Desktop",
		Family:  FamilyLinux,
		Version: "24.04.4 LTS",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://releases.ubuntu.com/24.04/ubuntu-24.04.4-desktop-amd64.iso", Size: "~6.0 GB"},
		},
	},
	{
		Name:    "Ubuntu Server",
		Family:  FamilyLinux,
		Version: "24.04.4 LTS",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso", Size: "~2.6 GB"},
			{Arch: ArchARM64, DownloadURL: "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.4-live-server-arm64.iso", Size: "~2.5 GB"},
		},
	},
	{
		Name:    "Fedora Workstation",
		Family:  FamilyLinux,
		Version: "42",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://download.fedoraproject.org/pub/fedora/linux/releases/42/Workstation/x86_64/iso/Fedora-Workstation-Live-x86_64-42-1.1.iso", Size: "~2.2 GB"},
			{Arch: ArchARM64, DownloadURL: "https://download.fedoraproject.org/pub/fedora/linux/releases/42/Workstation/aarch64/iso/Fedora-Workstation-Live-aarch64-42-1.1.iso", Size: "~2.2 GB"},
		},
	},
	{
		Name:    "Debian",
		Family:  FamilyLinux,
		Version: "13 (Trixie) netinst",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-13.0.0-amd64-netinst.iso", Size: "~650 MB"},
			{Arch: ArchARM64, DownloadURL: "https://cdimage.debian.org/debian-cd/current/arm64/iso-cd/debian-13.0.0-arm64-netinst.iso", Size: "~700 MB"},
			{Arch: ArchI386, DownloadURL: "https://cdimage.debian.org/debian-cd/current/i386/iso-cd/debian-13.0.0-i386-netinst.iso", Size: "~650 MB"},
		},
	},
	{
		Name:    "Linux Mint",
		Family:  FamilyLinux,
		Version: "22.1 Cinnamon",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://mirrors.kernel.org/linuxmint/stable/22.1/linuxmint-22.1-cinnamon-64bit.iso", Size: "~2.8 GB"},
		},
	},
	{
		Name:    "Arch Linux",
		Family:  FamilyLinux,
		Version: "2026.03.01",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://geo.mirror.pkgbuild.com/iso/2026.03.01/archlinux-2026.03.01-x86_64.iso", Size: "~1.1 GB"},
		},
	},
	{
		Name:    "openSUSE Leap",
		Family:  FamilyLinux,
		Version: "15.6",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://download.opensuse.org/distribution/leap/15.6/iso/openSUSE-Leap-15.6-DVD-x86_64-Current.iso", Size: "~4.4 GB"},
			{Arch: ArchARM64, DownloadURL: "https://download.opensuse.org/distribution/leap/15.6/iso/openSUSE-Leap-15.6-DVD-aarch64-Current.iso", Size: "~4.0 GB"},
		},
	},
	{
		Name:    "Manjaro",
		Family:  FamilyLinux,
		Version: "26.0.3 GNOME",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://download.manjaro.org/gnome/26.0.3/manjaro-gnome-26.0.3-260228-linux618.iso", Size: "~3.5 GB"},
		},
	},
	{
		Name:    "Pop!_OS",
		Family:  FamilyLinux,
		Version: "22.04 LTS",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://iso.pop-os.org/22.04/amd64/intel/46/pop-os_22.04_amd64_intel_46.iso", Size: "~2.5 GB"},
		},
	},
	{
		Name:    "Kali Linux",
		Family:  FamilyLinux,
		Version: "2026.1",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://cdimage.kali.org/kali-2026.1/kali-linux-2026.1-installer-amd64.iso", Size: "~4.1 GB"},
			{Arch: ArchARM64, DownloadURL: "https://cdimage.kali.org/kali-2026.1/kali-linux-2026.1-installer-arm64.iso", Size: "~3.9 GB"},
		},
	},
	{
		Name:    "Rocky Linux",
		Family:  FamilyLinux,
		Version: "9.7 Minimal",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://download.rockylinux.org/pub/rocky/9/isos/x86_64/Rocky-9.7-x86_64-minimal.iso", Size: "~1.8 GB"},
			{Arch: ArchARM64, DownloadURL: "https://download.rockylinux.org/pub/rocky/9/isos/aarch64/Rocky-9.7-aarch64-minimal.iso", Size: "~1.7 GB"},
		},
	},
	{
		Name:    "AlmaLinux",
		Family:  FamilyLinux,
		Version: "9.7 Minimal",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://repo.almalinux.org/almalinux/9/isos/x86_64/AlmaLinux-9.7-x86_64-minimal.iso", Size: "~1.8 GB"},
			{Arch: ArchARM64, DownloadURL: "https://repo.almalinux.org/almalinux/9/isos/aarch64/AlmaLinux-9.7-aarch64-minimal.iso", Size: "~1.7 GB"},
		},
	},

	// ── Windows ─────────────────────────────────────────────────
	{
		Name:    "Windows 11",
		Family:  FamilyWindows,
		Version: "24H2",
		Format:  "ISO (manual download)",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://www.microsoft.com/software-download/windows11", Size: "~5.4 GB"},
			{Arch: ArchARM64, DownloadURL: "https://www.microsoft.com/software-download/windows11arm64", Size: "~5.4 GB"},
		},
	},
	{
		Name:    "Windows 10",
		Family:  FamilyWindows,
		Version: "22H2",
		Format:  "ISO (manual download)",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://www.microsoft.com/software-download/windows10ISO", Size: "~5.0 GB"},
		},
	},
	{
		Name:    "Windows Server",
		Family:  FamilyWindows,
		Version: "2025 Eval",
		Format:  "ISO (eval)",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://www.microsoft.com/en-us/evalcenter/evaluate-windows-server-2025", Size: "~5.0 GB"},
		},
	},

	// ── Unix / BSD ──────────────────────────────────────────────
	{
		Name:    "FreeBSD",
		Family:  FamilyUnix,
		Version: "14.4",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://download.freebsd.org/releases/amd64/amd64/ISO-IMAGES/14.4/FreeBSD-14.4-RELEASE-amd64-disc1.iso", Size: "~1.1 GB"},
			{Arch: ArchARM64, DownloadURL: "https://download.freebsd.org/releases/arm64/aarch64/ISO-IMAGES/14.4/FreeBSD-14.4-RELEASE-arm64-aarch64-disc1.iso", Size: "~900 MB"},
			{Arch: ArchRISCV64, DownloadURL: "https://download.freebsd.org/releases/riscv/riscv64/ISO-IMAGES/14.4/FreeBSD-14.4-RELEASE-riscv-riscv64-disc1.iso", Size: "~800 MB"},
		},
	},
	{
		Name:    "OpenBSD",
		Family:  FamilyUnix,
		Version: "7.7",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://cdn.openbsd.org/pub/OpenBSD/7.7/amd64/install77.iso", Size: "~640 MB"},
			{Arch: ArchARM64, DownloadURL: "https://cdn.openbsd.org/pub/OpenBSD/7.7/arm64/install77.iso", Size: "~580 MB"},
			{Arch: ArchI386, DownloadURL: "https://cdn.openbsd.org/pub/OpenBSD/7.7/i386/install77.iso", Size: "~640 MB"},
			{Arch: ArchRISCV64, DownloadURL: "https://cdn.openbsd.org/pub/OpenBSD/7.7/riscv64/install77.iso", Size: "~580 MB"},
		},
	},
	{
		Name:    "NetBSD",
		Family:  FamilyUnix,
		Version: "10.1",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://cdn.netbsd.org/pub/NetBSD/NetBSD-10.1/images/NetBSD-10.1-amd64.iso", Size: "~2.5 GB"},
			{Arch: ArchI386, DownloadURL: "https://cdn.netbsd.org/pub/NetBSD/NetBSD-10.1/images/NetBSD-10.1-i386.iso", Size: "~2.3 GB"},
		},
	},
	{
		Name:    "DragonFly BSD",
		Family:  FamilyUnix,
		Version: "6.4.2",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://mirror-master.dragonflybsd.org/iso-images/dfly-x86_64-6.4.2_REL.iso", Size: "~1.1 GB"},
		},
	},
	{
		Name:    "GhostBSD",
		Family:  FamilyUnix,
		Version: "25.02",
		Format:  "ISO",
		Variants: []Variant{
			{Arch: ArchAMD64, DownloadURL: "https://download.ghostbsd.org/releases/amd64/latest/GhostBSD-25.02-R14.3p2.iso", Size: "~2.0 GB"},
		},
	},
}

func OSByFamily(family Family) []OS {
	var result []OS
	for _, os := range Catalog {
		if os.Family == family {
			result = append(result, os)
		}
	}
	return result
}
