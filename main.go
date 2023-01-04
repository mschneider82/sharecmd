package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mdp/qrterminal"
	"github.com/spf13/cast"
	"gopkg.in/alecthomas/kingpin.v2"
	"schneider.vip/share/clipboard"
	"schneider.vip/share/config"
	"schneider.vip/share/provider"
	"schneider.vip/share/provider/dropbox"
	"schneider.vip/share/provider/nextcloud"
	"schneider.vip/share/provider/opendrive"
	"schneider.vip/share/provider/seafile"
	"schneider.vip/share/urlshortener"
	"schneider.vip/share/urlshortener/biturl"
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
		case "opendrive":
			sharecmd.provider = opendrive.NewProvider(sharecmd.config.ProviderSettings["user"],
				sharecmd.config.ProviderSettings["pass"])
		case "dropbox":
			sharecmd.provider = dropbox.NewProvider(cfg.ProviderSettings["token"])
		case "nextcloud":
			sharecmd.provider = nextcloud.NewProvider(nextcloud.Config{
				URL:                   sharecmd.config.ProviderSettings["url"],
				Username:              sharecmd.config.ProviderSettings["username"],
				Password:              sharecmd.config.ProviderSettings["password"],
				LinkShareWithPassword: cast.ToBool(sharecmd.config.ProviderSettings["linkShareWithPassword"]),
				RandomPasswordChars:   cast.ToInt(sharecmd.config.ProviderSettings["randomPasswordChars"]),
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
		var qr strings.Builder
		qrterminal.Generate(link, qrterminal.L, &qr)
		fmt.Printf("\n%s\n", qr.String())
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
