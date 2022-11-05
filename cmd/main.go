package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
)

const (
	folder = `C:\ws\`
)

var (
	hostFlag    = flag.String("host", "", "The host which TXT record is set to an url of links")
	localSource = flag.String("local", "", "The local source of links")
	allFlag     = flag.Bool("all", false, "Download all links")
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {

	// temp := links{
	// 	Links: []link{
	// 		{
	// 			Url:     "https://download.sysinternals.com/files/SysinternalsSuite.zip",
	// 			Version: "latest",
	// 		},
	// 	},
	// 	LastModified: time.Now().UTC(),
	// }
	// d, _ := json.MarshalIndent(temp, "", "  ")
	// fmt.Println(string(d))

	flag.Parse()
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

	var links config.Links
	if *hostFlag != "" {
		l, err := getLinks(*hostFlag)
		if err != nil {
			fmt.Println(err)
			return
		}
		links = l
	} else {
		l, err := getLinksLocal(*localSource)
		if err != nil {
			fmt.Println(err)
			return
		}
		links = l
	}

	if *allFlag {
		err := createFolder(folder)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, l := range links.Links {
			err := downloader.Get(l, folder)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		return
	}

	fmt.Println(links)

	return
	fmt.Println("Please select table.")
	t := prompt.Input("> ", completer)
	fmt.Println("You selected " + t)
}

func createFolder(f string) error {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		err := os.Mkdir(f, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLinksLocal(f string) (config.Links, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return config.Links{}, err
	}

	var l config.Links
	err = json.Unmarshal(data, &l)
	if err != nil {
		return config.Links{}, err
	}
	return l, nil
}

func getLinks(host string) (config.Links, error) {
	txts, err := net.LookupTXT(host)
	if err != nil {
		return config.Links{}, err
	}

	if len(txts) == 0 {
		return config.Links{}, fmt.Errorf("no link")
	}

	if len(txts) > 1 {
		return config.Links{}, fmt.Errorf("too many links")
	}

	u := txts[0]
	if _, err := url.Parse(u); err != nil {
		return config.Links{}, err
	}

	resp, err := http.Get(u)
	if err != nil {
		return config.Links{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return config.Links{}, err
	}

	var l config.Links
	err = json.Unmarshal(body, &l)
	if err != nil {
		return config.Links{}, err
	}
	return l, nil
}
