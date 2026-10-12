package shipped

import (
	"fmt"
	"testing"

	"fixture/support"
)

func helperSpawns() <-chan int {
	c := make(chan int, 1)
	go func() { c <- 1 }()
	return c
}

type launcher interface{ launch() <-chan int }

type viaInterface struct{}

func (viaInterface) launch() <-chan int {
	c := make(chan int, 1)
	go func() { c <- 1 }()
	return c
}

func TestDirectGo(t *testing.T) {
	c := make(chan int)
	go func() { c <- 1 }()
	<-c
}

func TestGoInsideALiteral(t *testing.T) {
	t.Run("sub", func(t *testing.T) {
		c := make(chan int)
		go func() { c <- 1 }()
		<-c
	})
}

func TestCallsATestHelper(t *testing.T) { <-helperSpawns() }

func TestCallsASupportPackage(t *testing.T) { <-support.Launch() }

// Reaches a start only by calling into shipped source: NOT selected.
func TestCallsAShippedMethod(t *testing.T) { <-(&Svc{}).Start() }

// Reaches a start only by calling into shipped source: NOT selected.
func TestCallsAShippedFunction(t *testing.T) { <-Spawn() }

// Reaches a start only through a function value: NOT selected.
func TestThroughAFunctionValue(t *testing.T) {
	f := helperSpawns
	<-f()
}

// Reaches a start only through an interface: NOT selected.
func TestThroughAnInterface(t *testing.T) {
	var l launcher = viaInterface{}
	<-l.launch()
}

func TestPlain(t *testing.T) {
	if Plain() != 1 {
		t.Fatal("Plain")
	}
}

//racescope:covers shipped/shipped.go:Spawn
func TestOwnsSpawn(t *testing.T) { <-Spawn() }

//racescope:covers shipped/shipped.go:Svc.Start
func TestOwnsStart(t *testing.T) { <-(&Svc{}).Start() }

func FuzzSpawns(f *testing.F) {
	f.Add(1)
	f.Fuzz(func(t *testing.T, n int) {
		c := make(chan int)
		go func() { c <- n }()
		<-c
	})
}

func ExampleSpawn() {
	c := make(chan int)
	go func() { c <- 1 }()
	fmt.Println(<-c)
	// Output: 1
}
