package main

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	//app        = kingpin.New("store", "CLI Tool to share files (c) Matthias Schneider")
	configFile = kingpin.Flag("config", "Client configuration file").Default(userHomeDir() + "/.config/sharecmd/config.json").String()
	setup      = kingpin.Flag("setup", "Setup client configuration").Bool()

	file = kingpin.Arg("file", "filename to upload").File()

	appname = "gostore"
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

		fmt.Println("Got settings for provider: ", cfg.Provider)

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
