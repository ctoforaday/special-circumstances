// Package shipped is in the fixture binary's import graph.
package shipped

// Spawn starts a goroutine in shipped source.
func Spawn() <-chan int {
	c := make(chan int, 1)
	go func() { c <- 1 }()
	return c
}

type Svc struct{}

// Start starts a goroutine in a shipped-source method.
func (s *Svc) Start() <-chan int {
	c := make(chan int, 1)
	go func() { c <- 1 }()
	return c
}

// Plain starts nothing.
func Plain() int { return 1 }
