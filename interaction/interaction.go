package interaction

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
	"golang.org/x/exp/slices"
)

type interaction struct {
	lc config.LinksContainer
}

func Prompt(linksContainer config.LinksContainer) {
	interaction := interaction{
		lc: linksContainer,
	}

	fmt.Println("Please select file to download:")
	t := prompt.Input("> ", interaction.completer)

	if t == "" {
		fmt.Println("No file selected")
		return
	}

	i := slices.IndexFunc(linksContainer.Links, func(l config.Link) bool {
		return l.GetDisplayName() == t
	})

	if i == -1 {
		fmt.Println("No file found, please complete the whole name.")
		return
	}

	err := downloader.Get(linksContainer.Links[i], config.GetConfig().DestinationFolder)
	if err != nil {
		fmt.Println(err)
		return
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
