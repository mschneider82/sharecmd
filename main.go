package main

import (
	"log"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile = kingpin.Flag("config", "Client configuration file").Default(userHomeDir() + "/.config/sharecmd/config.json").String()
	setup      = kingpin.Flag("setup", "Setup client configuration").Bool()
	file       = kingpin.Arg("file", "filename to upload").File()
)

// ShareCmd cli app
type ShareCmd struct {
	config   *Config
	provider Provider
}

func main() {
	kingpin.Parse()

	if *setup {
		configSetup()
	}
	if file != nil {
		sharecmd := ShareCmd{}
		cfg, err := lookupConfig()
		if err != nil {
			log.Fatalf("lookupConfig: %v \n", err)
		}
		sharecmd.config = &cfg

		switch sharecmd.config.Provider {
		case "googledrive":
			sharecmd.provider = NewGoogleDriveProvider(sharecmd.config.ProviderSettings["googletoken"])
		case "dropbox":
			sharecmd.provider = NewDropboxProvider(cfg.ProviderSettings["token"])
		}

		fileid, err := sharecmd.provider.Upload(*file, "")
		if err != nil {
			log.Fatalf("Can't upload file: %s", err.Error())
		}
		link, err := sharecmd.provider.GetLink(fileid)
		if err != nil {
			log.Fatalf("Can't get link for file: %s", err.Error())
		}
		log.Printf("URL: %s", link)
		toClip(link)
	}
}
