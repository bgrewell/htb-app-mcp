// Command htb-app-mcp is an unofficial Model Context Protocol server for the
// HackTheBox main application. This is a Phase 0 stub: it prints version and
// exits. The real MCP bootstrap lands in Phase 1.
package main

import (
	"flag"
	"fmt"
	"os"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "0.0.0-dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	fmt.Fprintln(os.Stderr, "htb-app-mcp: pre-alpha; no domains enabled. See --help.")
	os.Exit(1)
}
