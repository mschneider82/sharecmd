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

func main() {
	kingpin.Parse()

	if *setup {
		configSetup()
	} else {

		cfg, err := lookupConfig()
		if err != nil {
			panic(fmt.Sprintf("lookupConfig: %v \n", err))
		}
		switch cfg.Provider {
		case "googledrive":
			gdrive := NewGoogleDriveProvider(cfg.ProviderSettings["googletoken"])
			fileid, err := gdrive.Upload(*file, "")
			if err != nil {
				panic(err)
			}
			if link, err := gdrive.GetLink(fileid); err == nil {
				fmt.Println(link)
			}
		case "dropbox":
			dbx := NewDropboxProvider(cfg.ProviderSettings["token"])
			dst, err := dbx.Upload(*file, "")
			if err != nil {
				panic(err)
			}
			if link, err := dbx.GetLink(dst); err == nil {
				fmt.Println(link)
			}
		}
	}
}
