package main

import (
	"fmt"

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
			panic(fmt.Sprintf("lookupConfig: %v \n", err))
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
			panic(err)
		}
		link, err := sharecmd.provider.GetLink(fileid)
		if err != nil {
			panic(err)
		}
		fmt.Println(link)

	}
}
