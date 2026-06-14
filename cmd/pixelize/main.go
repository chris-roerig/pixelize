package main

import (
	"fmt"
	"os"

	"github.com/chris-roerig/pixelize/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
