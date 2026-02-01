package main

import (
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		if exitErr, ok := err.(interface{ ExitCode() int }); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
