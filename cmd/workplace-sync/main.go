package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	update "github.com/dhcgn/gh-update"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
	"github.com/dhcgn/workplace-sync/interaction"
	"github.com/dhcgn/workplace-sync/linkscontainer"

	"github.com/pterm/pterm"
)

var (
	Version = "dev"
)

var (
	hostFlag    = flag.String("host", "", "The host which TXT record is set to an url of links")
	localSource = flag.String("local", "", "The local source of links")
	allFlag     = flag.Bool("all", false, "Download all links")
	nameFlag    = flag.String("name", "", "The name or preffix of the tool to download")
	updateFlag  = flag.Bool("update", false, "The name or preffix of the tool to download")
)

var (
	destFolder = config.GetConfig().DestinationFolder
)

func main() {
	fmt.Printf("Workplace Sync %v (%v %v)\n", Version, buildInfoTime, runtime.Version())
	fmt.Println("https://github.com/dhcgn/workplace-sync")
	fmt.Println()

	if update.IsFirstStartAfterUpdate() {
		fmt.Println("Update finished!")
		err := update.CleanUpAfterUpdate(os.Args[0])
		if err != nil {
			fmt.Println("ERROR Clean up:", err)
		}

		return
	}

	flag.Parse()

	if *updateFlag {
		fmt.Print("Checking for updates ... ")
		err := update.SelfUpdateWithLatestAndRestart("dhcgn/workplace-sync", Version, "^ws-.*windows.*zip$", os.Args[0])

		if err != nil && err == update.ErrorNoNewVersionFound {
			fmt.Println("No new version found")
		} else if err != nil {
			fmt.Println("ERROR Update:", err)
		}

		return
	}

	if *hostFlag == "" && *localSource == "" {
		fmt.Println("host or localSource is required")
		flag.PrintDefaults()
		return
	}

	if *hostFlag != "" && *localSource != "" {
		fmt.Println("host and localSource are mutually exclusive")
		flag.PrintDefaults()
		return
	}

	if *allFlag && *nameFlag != "" {
		fmt.Println("all and name are mutually exclusive")
		flag.PrintDefaults()
		return
	}

	var linksContainer config.LinksContainer
	if *hostFlag != "" {
		pterm.Info.Printfln("Optain links from DNS TXT record of %v", *hostFlag)
		l, err := linkscontainer.GetLinksDNS(*hostFlag)
		if err != nil {
			fmt.Println(err)
			return
		}
		linksContainer = l
	} else {
		pterm.Info.Printfln("Optain links from local file %v", *localSource)
		l, err := linkscontainer.GetLinksLocal(*localSource)
		if err != nil {
			fmt.Println(err)
			return
		}
		linksContainer = l
	}

	pterm.Success.Printfln("Got %v links", len(linksContainer.Links))

	pterm.Info.Printfln("Use download folder %v", destFolder)
	err := createDownloadFolder(destFolder)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *allFlag {
		pterm.Info.Printfln("All links will be downloaded:")
		for i, l := range linksContainer.Links {
			pterm.Info.Printfln("%2v. %v (%v)", (i + 1), l.GetDisplayName(), l.Version)
		}

		for _, l := range linksContainer.Links {
			err := downloader.Get(l, destFolder)
			if err != nil {
				pterm.Error.Printfln("link %v, folder: %v, error: %v", l.Url, destFolder, err)
				continue
			}
		}
		return
	}

	if *nameFlag != "" {
		interaction.Download(*nameFlag, linksContainer)
		return
	}

	interaction.PromptAndDownload(linksContainer)

	fmt.Println()
	fmt.Println("Done")
}

func createDownloadFolder(f string) error {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		err := os.Mkdir(f, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
