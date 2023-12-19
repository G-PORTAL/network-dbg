package main

import (
	"flag"
	"fmt"
	"github.com/G-PORTAL/network-dbg/pkg/locations"
	"github.com/G-PORTAL/network-dbg/pkg/traceroute"
	"github.com/charmbracelet/log"
	"runtime"
)

var locationLimit *string
var locationShow *bool
var showHelp *bool

func main() {
	if *showHelp {
		log.Info("GPORTAL Network Debug Tool")
		flag.Usage()
		return
	}

	endpoints := locations.GetLocations(locationLimit)
	if *locationShow {
		log.Infof("Showing %d GPORTAL locations...", len(endpoints))
		for _, endpoint := range endpoints {
			log.Printf("- %s", endpoint.Short())
		}
		return
	}

	log.Infof("Testing %d GPORTAL locations now...", len(endpoints))

	traceroute.Run(endpoints)

	// Kepp the window open on windows
	if runtime.GOOS == "windows" {
		log.Info("Press any key to continue...")
		var input string
		_, _ = fmt.Scanln(&input)
	}
}

func init() {
	showHelp = flag.Bool("help", false, "Show this help message")
	locationShow = flag.Bool("show", false, "Print all available location endpoints")
	locationLimit = flag.String("limit", "", "Limit to a specific location (example: fra01)")

	flag.Parse()
}
