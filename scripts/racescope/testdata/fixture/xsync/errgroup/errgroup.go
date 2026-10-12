// Package errgroup stands in for golang.org/x/sync/errgroup at the same import path, so the
// fixture resolves the table's row without a download.
package errgroup

import "sync"

type Group struct{ wg sync.WaitGroup }

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		_ = f()
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	return nil
}
