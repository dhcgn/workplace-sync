package interaction

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
	"github.com/pterm/pterm"
)

type interaction struct {
	lc config.LinksContainer
}

func Download(t string, linksContainer config.LinksContainer) {
	if t == "" {
		pterm.Error.Printfln("No file selected")
		return
	}

	found := make([]config.Link, 0)
	for _, l := range linksContainer.Links {
		if l.GetDisplayName() == t {
			found = append(found, l)
		}
	}

	if len(found) > 1 {
		pterm.Error.Printfln("Multiple matches found for: %v", t)
		return
	}

	if len(found) == 1 {
		sum, err := downloader.DownloadFileToFolder(found[0], config.GetConfig().DestinationFolder)
		if err != nil {
			pterm.Error.Println(err)
			return
		}
		if found[0].Hash == "" {
			PrintHashWarning(found[0].GetDisplayName(), sum)
		}
		return
	}

	pterm.Warning.Printfln("No file found, try case-ignore prefix")

	match := linksContainer.GetLinksByDisplayNamePreffix(t)

	if len(match) == 0 {
		pterm.Error.Printfln("No file found with suffix: %v", t)
		return
	}

	if len(match) != 1 {
		pterm.Error.Printfln(`Multiple files found with suffix: %v`, t)
		for _, l := range match {
			pterm.Error.Printfln(`Found: %v`, l.GetDisplayName())
		}
		return
	}

	pterm.Success.Printfln("Found file %v", match[0].GetDisplayName())

	sum, err := downloader.DownloadFileToFolder(match[0], config.GetConfig().DestinationFolder)
	if err != nil {
		pterm.Error.Print(err)
	}
	if match[0].Hash == "" {
		PrintHashWarning(match[0].GetDisplayName(), sum)
	}
}

func PrintHashWarning(name string, hash string) {
	pterm.Warning.Printfln("No hash check for %v, downloaded content hash was %v", name, hash)
}

func PromptAndDownload(linksContainer config.LinksContainer) {
	interaction := interaction{
		lc: linksContainer,
	}

	pterm.Info.Printfln("Please select file to download:")
	t := prompt.Input("> ", interaction.completer, prompt.OptionMaxSuggestion(20))

	Download(t, linksContainer)
}

func (i interaction) completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest
	for _, l := range i.lc.Links {
		s := prompt.Suggest{
			Text:        l.GetDisplayName(),
			Description: fmt.Sprintf("%v from %v", l.Version, l.GetHostFromLink()),
		}
		suggestions = append(suggestions, s)
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}
