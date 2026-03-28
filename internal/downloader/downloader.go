package downloader

import (
	"github.com/cavaliergopher/grab/v3"
)

type Downloader struct {
	client *grab.Client
	req    *grab.Request
	resp   *grab.Response
}

func New() *Downloader {
	return &Downloader{
		client: grab.NewClient(),
	}
}

func (d *Downloader) Start(url string, dest string) error {
	req, err := grab.NewRequest(dest, url)
	if err != nil {
		return err
	}
	d.req = req
	d.resp = d.client.Do(req)
	return d.resp.Err()
}

func (d *Downloader) Progress() float64 {
	if d.resp == nil {
		return 0
	}
	return d.resp.Progress()
}

func (d *Downloader) IsComplete() bool {
	if d.resp == nil {
		return false
	}
	return d.resp.IsComplete()
}

func (d *Downloader) Err() error {
	if d.resp == nil {
		return nil
	}
	return d.resp.Err()
}

func (d *Downloader) Wait() error {
	if d.resp == nil {
		return nil
	}
	<-d.resp.Done
	return d.resp.Err()
}
