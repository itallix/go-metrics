package main

import (
	"os"
)

func main() {
	os.Exit(1) // want "usage of os.Exit is not allowed in the main function of the main package"
}

func osExitCheckFunc() {
	os.Exit(1)
}
