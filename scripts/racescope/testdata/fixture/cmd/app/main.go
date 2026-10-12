// Command app is the fixture's shipped binary: its import graph is the race scope.
package main

import "fixture/shipped"

func main() { <-shipped.Spawn() }
