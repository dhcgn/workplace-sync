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
	hostFlag        = flag.String("host", "", "The host which DNS TXT record points to an url of links.json")
	localSource     = flag.String("local", "", "The local source of links (.json)")
	allFlag         = flag.Bool("all", false, "Download all links, except skipped ones")
	nameFlag        = flag.String("name", "", "The name or preffix of the tool to download")
	updateFlag      = flag.Bool("update", false, "Update app with latest release from github.com")
	checkUpdateFlag = flag.Bool("checkupdate", false, "Check for update from github.com")
	versionFlag     = flag.Bool("version", false, "Return version of app")
)

var (
	destFolder     = config.GetConfig().DestinationFolder
	forceHashCheck = config.GetConfig().ForceHashCheck
)

const (
	updateName       = "dhcgn/workplace-sync"
	updateAssetRegex = "^ws-.*windows.*zip$"
)

func main() {
	fmt.Printf("Workplace Sync %v (%v %v)\n", Version, buildInfoTime, runtime.Version())
	fmt.Println("https://github.com/dhcgn/workplace-sync")
	fmt.Println()

	if update.IsFirstStartAfterUpdate() {
		fmt.Println("Update finished!")
		oldPid := update.GetOldPid()
		if oldPid != fmt.Sprint(os.Getpid()) {
			err := update.CleanUpAfterUpdate(os.Args[0], oldPid)
			if err != nil {
				fmt.Println("ERROR Clean up:", err)
			}
		} else {
			fmt.Println("ERROR: PID is the same!")
		}

		return
	}

	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		return
	}

	if *checkUpdateFlag {
		lr, err := update.GetLatestVersion(updateName, Version, updateAssetRegex)
		if err != nil && err == update.ErrorNoNewVersionFound {
			fmt.Printf("No new version found for %v\n", Version)
			return
		} else if err != nil {
			fmt.Println("ERROR Update:", err)
			return
		}

		fmt.Printf("Could update from %v to %v, run '-update' to update this app.\n", Version, lr.Version)

		return
	}

	if *updateFlag {
		fmt.Println("Checking for updates ... ")

		lr, err := update.GetLatestVersion(updateName, Version, updateAssetRegex)
		if err != nil && err == update.ErrorNoNewVersionFound {
			fmt.Printf("No new version found for %v\n", Version)
			return
		} else if err != nil {
			fmt.Println("ERROR Update:", err)
			return
		}

		fmt.Printf("Update %v to %v\n", Version, lr.Version)
		fmt.Println("Downloading update ... ")

		err = update.SelfUpdateAndRestart(lr, os.Args[0])

		if err != nil {
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
		pterm.Info.Printfln("Obtain links from DNS TXT record of %v", *hostFlag)
		l, err := linkscontainer.GetLinksDNS(*hostFlag)
		if err != nil {
			fmt.Println(err)
			return
		}
		linksContainer = l
	} else {
		pterm.Info.Printfln("Obtain links from local file %v", *localSource)
		l, err := linkscontainer.GetLinksLocal(*localSource)
		if err != nil {
			fmt.Println(err)
			return
		}
		linksContainer = l
	}

	if forceHashCheck {
		var filteredLinksContainer config.LinksContainer
		for _, l := range linksContainer.Links {
			if l.Hash == "" {
				pterm.Warning.Printfln("Link %v (%v) has no hash, will be removed because ForceHashCheck is active", l.GetDisplayName(), l.Version)
			} else {
				filteredLinksContainer.Links = append(filteredLinksContainer.Links, l)
			}
		}
		linksContainer = filteredLinksContainer
	}

	if len(linksContainer.Links) == 0 {
		pterm.Error.Printfln("No links found")
		return
	} else {
		pterm.Success.Printfln("Got %v links", len(linksContainer.Links))
	}

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

			if l.Skipped {
				pterm.Info.Printfln("Skipped %v (%v)", l.GetDisplayName(), l.Version)
				continue
			}

			hash, err := downloader.Get(l, destFolder)
			if err != nil {
				pterm.Error.Printfln("link %v, folder: %v, error: %v", l.Url, destFolder, err)
				continue
			}
			if l.Hash == "" {
				interaction.PrintHashWarning(l.GetDisplayName(), hash)
			}
		}
		return
	}

	if *nameFlag != "" {
		interaction.Download(*nameFlag, linksContainer)
		return
	}

	pterm.Info.Printfln("The following tools are available:")
	for i, l := range linksContainer.Links {
		fmt.Printf("%v (%v)", l.GetDisplayName(), l.Version)
		if i != len(linksContainer.Links)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()

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
