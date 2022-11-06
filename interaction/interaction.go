package interaction

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
	"github.com/pterm/pterm"
	"golang.org/x/exp/slices"
)

type interaction struct {
	lc config.LinksContainer
}

func Prompt(linksContainer config.LinksContainer) {
	interaction := interaction{
		lc: linksContainer,
	}

	pterm.Info.Printfln("Please select file to download:")
	t := prompt.Input("> ", interaction.completer)

	if t == "" {
		pterm.Error.Printfln("No file selected")
		return
	}

	i := slices.IndexFunc(linksContainer.Links, func(l config.Link) bool {
		return l.GetDisplayName() == t
	})

	if i != -1 {
		err := downloader.Get(linksContainer.Links[i], config.GetConfig().DestinationFolder)
		if err != nil {
			pterm.Error.Print(err)
		}
		return
	}

	pterm.Warning.Printfln("No file found, try case-ignore prefix")

	var match []config.Link
	for _, l := range linksContainer.Links {
		if strings.HasPrefix(strings.ToLower(l.GetDisplayName()), strings.ToLower(t)) {
			match = append(match, l)
		}
	}

	if len(match) == 0 {
		pterm.Error.Printfln("No file found with suffix %v", t)
		return
	}

	if len(match) != 1 {
		pterm.Error.Printfln("Multiple files found with suffix %v", t)
		return
	}

	pterm.Success.Printfln("Found file %v", match[0].GetDisplayName())

	err := downloader.Get(match[0], config.GetConfig().DestinationFolder)
	if err != nil {
		pterm.Error.Print(err)
	}
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
