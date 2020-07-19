package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/netraitcorp/netick/pkg/server"
)

const (
	appName    = "netick"
	appDesc    = ""
	appVersion = "0.1.0"
)

var (
	buildTime string
	goVersion string
)

var usages = fmt.Sprintf(`
Usage: %s [OPTIONS]

%s

Options:
  -a, --addr <host>        Server running address (default: 0.0.0.0:2634)
  -v, --version            Show version
  -h, --help               Show this help
`, appName, appDesc)

var (
	showHelpFlag    bool
	showVersionFlag bool
	addressFlag     string
)

func usage() {
	fmt.Printf("%s\n", usages)
}

func parseFlags() error {
	usaf := flag.NewFlagSet(appName, flag.ContinueOnError)
	usaf.Usage = usage

	usaf.BoolVar(&showHelpFlag, "help", false, "Show this help")
	usaf.BoolVar(&showHelpFlag, "h", false, "Show this help")
	usaf.BoolVar(&showVersionFlag, "version", false, "Show version")
	usaf.BoolVar(&showVersionFlag, "v", false, "Show version")
	usaf.StringVar(&addressFlag, "a", "0.0.0.0:2634", "Server running address")

	if err := usaf.Parse(os.Args[1:]); err != nil {
		return err
	}

	if showHelpFlag {
		usaf.Usage()
		os.Exit(0)
	}

	if showVersionFlag {
		fmt.Printf("%s version %s\n", appName, appVersion)
		if goVersion != "" && buildTime != "" {
			fmt.Printf("built by %s, %s\n", goVersion, buildTime)
		}
		os.Exit(0)
	}

	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		os.Exit(2)
	}

	go func() {
		_ = http.ListenAndServe("", nil)
	}()

	if err := server.RunServer(addressFlag); err != nil {
		log.Fatalf("server run error: %s\n", err.Error())
	}
}
