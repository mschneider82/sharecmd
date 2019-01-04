package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mschneider82/sharecmd/provider/nextcloud"
	"github.com/mschneider82/sharecmd/provider/seafile"
	"github.com/mschneider82/sharecmd/urlshortener"
	"github.com/mschneider82/sharecmd/urlshortener/biturl"

	"github.com/mschneider82/sharecmd/clipboard"
	"github.com/mschneider82/sharecmd/config"
	"github.com/mschneider82/sharecmd/provider"
	"github.com/mschneider82/sharecmd/provider/dropbox"
	"github.com/mschneider82/sharecmd/provider/googledrive"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile  = kingpin.Flag("config", "Client configuration file").Default(config.UserHomeDir() + "/.config/sharecmd/config.json").String()
	setup       = kingpin.Flag("setup", "Setup client configuration").Bool()
	file        = kingpin.Arg("file", "filename to upload").File()
	versionflag = kingpin.Flag("version", "print build Version").Short('v').Bool()
	version     = "0.0.0"
)

// ShareCmd cli app
type ShareCmd struct {
	config   *config.Config
	provider provider.Provider
	shorturl urlshortener.URLShortener
}

func main() {
	kingpin.Parse()

	if *versionflag {
		fmt.Printf("ShareCmd Version: %s\n", version)
		os.Exit(0)
	}

	if *setup {
		config.Setup(*configFile)
		os.Exit(0)
	}
	if file != nil {
		sharecmd := ShareCmd{}
		cfg, err := config.LookupConfig(*configFile)
		if err != nil {
			log.Fatalf("lookupConfig: %v \n", err)
		}
		sharecmd.config = &cfg

		switch sharecmd.config.Provider {
		case "seafile":
			sharecmd.provider = seafile.NewProvider(sharecmd.config.ProviderSettings["url"], sharecmd.config.ProviderSettings["token"], sharecmd.config.ProviderSettings["repoid"])
		case "googledrive":
			sharecmd.provider = googledrive.NewProvider(sharecmd.config.ProviderSettings["googletoken"])
		case "dropbox":
			sharecmd.provider = dropbox.NewProvider(cfg.ProviderSettings["token"])
		case "nextcloud":
			sharecmd.provider = nextcloud.NewProvider(nextcloud.Config{
				URL:      sharecmd.config.ProviderSettings["url"],
				Username: sharecmd.config.ProviderSettings["username"],
				Password: sharecmd.config.ProviderSettings["password"],
			})
		default:
			config.Setup(*configFile)
			os.Exit(0)
		}

		fileid, err := sharecmd.provider.Upload(*file, "")
		if err != nil {
			log.Fatalf("Can't upload file: %s", err.Error())
		}
		link, err := sharecmd.provider.GetLink(fileid)
		if err != nil {
			log.Fatalf("Can't get link for file: %s", err.Error())
		}
		fmt.Printf("URL: %s\n", link)
		switch sharecmd.config.URLShortenerProvider {
		case "biturl":
			sharecmd.shorturl = biturl.New(link)
			shorturl, err := sharecmd.shorturl.ShortURL()
			if err == nil {
				link = shorturl
				fmt.Printf("Short URL: %s\n", link)
			}
		default:
		}
		clipboard.ToClip(link)
	}
}
