package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/netraitcorp/netick/pkg/types"

	"github.com/netraitcorp/netick/pkg/log"
	"github.com/netraitcorp/netick/pkg/server"
)

const (
	appName    = "netick"
	appDesc    = "A simple and high performance open source messaging system for web application"
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
  -c, --config <file>      Configuration file (default: ./netick.yaml)
  --dev                    Starts the server in development mode
  -v, --version            Show version
  -h, --help               Show this help
`, appName, appDesc)

var (
	showHelpFlag    bool
	showVersionFlag bool
	addressFlag     string
	configFlag      string
	envDevelopFlag  bool
)

func usage() {
	fmt.Printf("%s\n", usages)
}

func parseFlags() error {
	usaf := flag.NewFlagSet(appName, flag.ContinueOnError)
	usaf.Usage = usage

	usaf.StringVar(&addressFlag, "a", "0.0.0.0:2634", "Server running address")
	usaf.StringVar(&addressFlag, "addr", "0.0.0.0:2634", "Server running address")
	usaf.StringVar(&configFlag, "config", "./netick.yaml", "Configuration file")
	usaf.StringVar(&configFlag, "c", "./netick.yaml", "Configuration file")
	usaf.BoolVar(&showHelpFlag, "help", false, "Show this help")
	usaf.BoolVar(&showHelpFlag, "h", false, "Show this help")
	usaf.BoolVar(&showVersionFlag, "version", false, "Show version")
	usaf.BoolVar(&showVersionFlag, "v", false, "Show version")
	usaf.BoolVar(&envDevelopFlag, "dev", false, "Starts the server in development mode")

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

	welcome()

	return nil
}

func welcome() {
	fmt.Println("   _  __      __   _       __")
	fmt.Println("  / |/ /___  / /_ (_)____ / /__")
	fmt.Println(" /    // -_)/ __// // __//  '_/")
	fmt.Println("/_/|_/ \\__/ \\__//_/ \\__//_/\\_\\")
	fmt.Println("")
	log.StdInfo("Version is %s", appVersion)
	log.StdInfo("Configuration loaded from file %s", configFlag)
	log.StdInfo("Started Websocket Server on %s", addressFlag)
	if envDevelopFlag {
		log.StdInfo("Starts the server in development mode")
	}
	log.StdInfo("Server is ready")
}

func main() {
	if err := parseFlags(); err != nil {
		os.Exit(2)
	}

	configOpts := log.NewOptions()
	if envDevelopFlag {
		configOpts.Env = types.EnvDev
	}
	configOpts.Level = "debug"
	log.InitLogger(configOpts)

	srvOpts := server.NewOptions()
	srvOpts.WebsocketOpts.Addr = addressFlag

	if err := server.RunWebsocketServer(srvOpts); err != nil {
		log.Fatal("tcp server run error: %s\n", err.Error())
	}
}
