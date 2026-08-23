package main

import (
	"fmt"
	"os"

	"github.com/amirmohammad/shellector/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "shellector:", err)
		os.Exit(1)
	}
}
