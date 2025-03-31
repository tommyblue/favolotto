//go:build !pn7150

package pn7150

import "context"

type Dummy struct{}

func New() (*Dummy, error) {
	return &Dummy{}, nil
}

func (d *Dummy) Run(ctx context.Context) error {
	return nil
}

func (d *Dummy) Stop() error {
	return nil
}

func (d *Dummy) Read() <-chan string {
	return make(chan string)
}
