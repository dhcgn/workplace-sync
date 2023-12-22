package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	update "github.com/dhcgn/gh-update"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
	"github.com/dhcgn/workplace-sync/interaction"
	"github.com/dhcgn/workplace-sync/linkscontainer"
	"golang.org/x/mod/semver"

	"github.com/pterm/pterm"
)

var (
	Version = "dev"
)

var (
	hostFlag        = flag.String("host", "", "The host which DNS TXT record points to an url of links.json")
	localSource     = flag.String("local", "", "The local source of links (.json)")
	urlSource       = flag.String("url", "", "The url of links.json")
	allFlag         = flag.Bool("all", false, "Download all links, except skipped ones")
	nameFlag        = flag.String("name", "", "The name or preffix of the tool to download")
	checkLinksFlag  = flag.Bool("checklinks", false, "Check if all links responds with 200")
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

	if *hostFlag == "" && *localSource == "" && *urlSource == "" {
		fmt.Println("host, urlSource or localSource is required")
		flag.PrintDefaults()
		return
	}

	notEmptyCount := 0
	if *hostFlag != "" {
		notEmptyCount++
	}
	if *localSource != "" {
		notEmptyCount++
	}
	if *urlSource != "" {
		notEmptyCount++
	}

	// host, localSource and urlSource are mutually exclusive, only one can be set
	if notEmptyCount > 1 {
		fmt.Println("host, localSource and urlSource are mutually exclusive")
		flag.PrintDefaults()
		return
	}

	if *allFlag && *nameFlag != "" {
		fmt.Println("all and name are mutually exclusive")
		flag.PrintDefaults()
		return
	}

	linksContainer, err := getLinksContainer(*hostFlag, *localSource, *urlSource)
	if err != nil {
		pterm.Error.Printfln("Error getting links: %v", err)
		return
	}

	checkMinVersion(linksContainer.MinVersion, Version)

	infos, warns, errors, err := checkAndFilterLinksContainer(&linksContainer, forceHashCheck, *checkLinksFlag)
	if err != nil {
		pterm.Error.Printfln("Error checking links: %v", err)
		return
	}
	for _, i := range infos {
		pterm.Info.Printfln(i)
	}
	for _, w := range warns {
		pterm.Warning.Printfln(w)
	}
	for _, w := range errors {
		pterm.Error.Printfln(w)
	}

	if len(linksContainer.Links) == 0 {
		pterm.Error.Printfln("No links found")
		return
	} else {
		pterm.Success.Printfln("Got %v links", len(linksContainer.Links))
	}

	pterm.Info.Printfln("Use download folder %v", destFolder)
	err = createDownloadFolder(destFolder)
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

func checkMinVersion(minVersion, currentVersion string) {
	if minVersion == "" || minVersion == "dev" {
		return
	}

	if !semver.IsValid(minVersion) {
		pterm.Error.Printfln("Remote repo has invalid min_version: %v. File could be incompatible.", minVersion)
		return
	}

	if !semver.IsValid(currentVersion) {
		pterm.Error.Printfln("App has invalid version: %v", currentVersion)
		return
	}

	if semver.Compare(minVersion, currentVersion) > 0 {
		pterm.Error.Printfln("Current version %v is lower than remote repo min_version %v. File could be incompatible.", currentVersion, minVersion)
	}
}

var (
	checkLinkFunc = func(url string) error {
		r, err := http.Head(url)
		if err != nil {
			return err
		}
		if r.StatusCode != 200 {
			return fmt.Errorf("Status code %v", r.StatusCode)
		}
		return nil
	}
)

func checkAndFilterLinksContainer(lc *config.LinksContainer, forceHashCheck, checkLink bool) (infos, warnings, errors []string, err error) {
	if forceHashCheck {
		var filteredLinksContainer config.LinksContainer
		for _, l := range lc.Links {
			if l.Hash == "" {
				warnings = append(warnings, fmt.Sprintf("Link %v (%v) has no hash, will be removed because ForceHashCheck is active", l.GetDisplayName(), l.Version))
			} else {
				filteredLinksContainer.Links = append(filteredLinksContainer.Links, l)
			}
		}
		lc = &filteredLinksContainer
	}

	var filteredLinksContainer config.LinksContainer
	for _, l := range lc.Links {
		if l.GithubReleaseAssetFilter != "" && ((l.Version != "" && l.Version != "latest") || l.Hash != "") {
			warnings = append(warnings, fmt.Sprintf("Link %v (%v) has GithubReleaseAssetFilter so and Version or Hash will be ignored", l.GetDisplayName(), l.Version))
		}
		if strings.HasPrefix(l.Url, "https://github.com/") && l.GithubReleaseAssetFilter == "" {
			warnings = append(warnings, fmt.Sprintf("Link %v (%v) has no GithubReleaseAssetFilter, so download is pinned and not latest release is used", l.GetDisplayName(), l.Version))
		}

		if checkLink {
			err := checkLinkFunc(l.Url)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Link %v (%v) is invalid: %v, url: %v", l.GetDisplayName(), l.Version, err, l.Url))
			} else {
				infos = append(infos, fmt.Sprintf("Link %v (%v) is valid, url: %v", l.GetDisplayName(), l.Version, l.Url))
				filteredLinksContainer.Links = append(filteredLinksContainer.Links, l)
			}
		}
	}
	lc = &filteredLinksContainer

	return infos, warnings, errors, nil
}

func getLinksContainer(host, path, url string) (config.LinksContainer, error) {
	if host != "" {
		pterm.Info.Printfln("Obtain links from DNS TXT record of %v", host)
		l, err := linkscontainer.GetLinksDNS(host)
		if err != nil {
			fmt.Println(err)
			return config.LinksContainer{}, err
		}
		return l, nil
	}

	if path != "" {
		pterm.Info.Printfln("Obtain links from local file %v", path)
		l, err := linkscontainer.GetLinksLocal(path)
		if err != nil {
			return config.LinksContainer{}, err
		}
		return l, nil
	}

	if url != "" {
		pterm.Info.Printfln("Obtain links from url %v", url)
		l, err := linkscontainer.GetLinksURL(url)
		if err != nil {
			return config.LinksContainer{}, err
		}
		return l, nil
	}
	return config.LinksContainer{}, fmt.Errorf("host, path or url is required")
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
