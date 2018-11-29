package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/mschneider82/sharecmd/provider/dropbox"
	"github.com/mschneider82/sharecmd/provider/googledrive"
	"golang.org/x/oauth2"
)

var providers = []string{"dropbox", "googledrive"}

// Config File Structure
type Config struct {
	Provider         string            `json:"provider"`
	ProviderSettings map[string]string `json:"providersettings"`
	Path             string
}

// UserHomeDir
func UserHomeDir() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env)
}

// Write config to disk
func (c Config) Write() error {
	err := os.MkdirAll(path.Dir(c.Path), 0700)
	if err != nil {
		return err
	}
	fmt.Printf("Saving config to %s \n", c.Path)
	p := c.Path
	c.Path = ""
	output, err := os.Create(p)
	if err != nil {
		return err
	}
	defer output.Close()

	return json.NewEncoder(output).Encode(c)
}

// LoadConfig from disk
func LoadConfig(path string) (Config, error) {
	config := Config{
		Path:             path,
		ProviderSettings: make(map[string]string),
	}

	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return config, nil
	} else if err != nil {
		return config, err
	}

	err = json.Unmarshal(content, &config)
	config.Path = path

	return config, err
}

// LookupConfig search config and load it
func LookupConfig(configfilepath string) (Config, error) {
	if configfilepath == "" {
		configfilepath = UserHomeDir() + "/.config/sharecmd/config.json"
	}

	config, err := LoadConfig(configfilepath)
	return config, err
}

// Setup asks user for input
func Setup(configfilepath string) error {
	config, err := LookupConfig(configfilepath)
	if err != nil {
		return err
	}

	prompt := promptui.Select{
		Label: "Select Provider",
		Items: providers,
	}

	_, provider, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	config.Provider = provider
	fmt.Printf("You choose %q\n", provider)

	switch provider {
	case "googledrive":
		conf := googledrive.OAuth2GoogleDriveConfig()
		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")

		authorizationprompt := promptui.Prompt{
			Label:   "Authorization Code",
			Default: "",
		}
		authcode, err := authorizationprompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		tok, err := conf.Exchange(context.TODO(), authcode)
		if err != nil {
			log.Fatalf("Unable to retrieve token from web %v", err)
		}

		tokenB, err := json.Marshal(tok)
		if err != nil {
			log.Fatalf("Unable to marshal json token %v", err)
		}

		config.ProviderSettings["googletoken"] = string(tokenB)

	case "dropbox":
		conf := dropbox.OAuth2DropboxConfig()
		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")

		authorizationprompt := promptui.Prompt{
			Label:   "Authorization Code",
			Default: config.ProviderSettings["token"],
		}
		authcode, err := authorizationprompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		var token *oauth2.Token
		ctx := context.Background()
		token, err = conf.Exchange(ctx, authcode)
		if err != nil {
			return err
		}
		config.ProviderSettings["token"] = token.AccessToken
	}

	return config.Write()
}
