package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if code := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}
func usage(w *os.File) { fmt.Fprintln(w, "usage: skilltrace source <scan|health> [options]") }
