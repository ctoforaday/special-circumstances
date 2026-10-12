// Package support is test support: no shipped package imports it, so it is outside the scope.
package support

// Launch starts a goroutine in a support package.
func Launch() <-chan int {
	c := make(chan int, 1)
	go func() { c <- 1 }()
	return c
}
