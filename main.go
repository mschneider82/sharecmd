package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mschneider82/sharecmd/clipboard"
	"github.com/mschneider82/sharecmd/config"
	"github.com/mschneider82/sharecmd/provider"
	"github.com/mschneider82/sharecmd/provider/dropbox"
	"github.com/mschneider82/sharecmd/provider/googledrive"
	"github.com/mschneider82/sharecmd/provider/seafile"
	"go.uber.org/dig"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile = kingpin.Flag("config", "Client configuration file").Default(config.UserHomeDir() + "/.config/sharecmd/config.json").String()
	setup      = kingpin.Flag("setup", "Setup client configuration").Bool()
	file       = kingpin.Arg("file", "filename to upload").File()
)

// ShareCmd cli app
type ShareCmd struct {
	config *config.Config
}

func main() {
	kingpin.Parse()
	c := dig.New()

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
			err = c.Provide(func() provider.Provider {
				return seafile.NewProvider(sharecmd.config.ProviderSettings["url"], sharecmd.config.ProviderSettings["token"], sharecmd.config.ProviderSettings["repoid"])
			})
		case "googledrive":
			err = c.Provide(func() provider.Provider {
				return googledrive.NewProvider(sharecmd.config.ProviderSettings["googletoken"])
			})
		case "dropbox":
			err = c.Provide(func() provider.Provider {
				return dropbox.NewProvider(cfg.ProviderSettings["token"])
			})
		default:
			config.Setup(*configFile)
			os.Exit(0)
		}
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		if err := c.Invoke(func(p provider.Provider) {
			fileid, err := p.Upload(*file, "")
			if err != nil {
				log.Fatalf("Can't upload file: %s", err.Error())
			}
			link, err := p.GetLink(fileid)
			if err != nil {
				log.Fatalf("Can't get link for file: %s", err.Error())
			}
			fmt.Printf("URL: %s\n", link)
			clipboard.ToClip(link)
		}); err != nil {
			log.Fatalf("%s", err.Error())
		}
	}
}
