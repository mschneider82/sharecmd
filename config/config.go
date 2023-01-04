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
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"
	"schneider.vip/share/provider/dropbox"
	"schneider.vip/share/provider/googledrive"
	"schneider.vip/share/provider/nextcloud"
	"schneider.vip/share/provider/seafile"
	"schneider.vip/share/urlshortener"
)

var providers = []string{"dropbox", "opendrive", "seafile", "nextcloud"}

// Config File Structure
type Config struct {
	Provider             string            `json:"provider"`
	ProviderSettings     map[string]string `json:"providersettings"`
	Path                 string
	URLShortenerProvider string
	URLShortenerSettings map[string]string
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
	err := os.MkdirAll(path.Dir(UserHomeDir()+"/.config/sharecmd/config.json"), 0o700)
	if err != nil {
		return err
	}
	fmt.Printf("Saving config to %s \n", UserHomeDir()+"/.config/sharecmd/config.json")
	p := UserHomeDir() + "/.config/sharecmd/config.json"
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
	config := Config{
		Path:             "/sharecmd",
		ProviderSettings: make(map[string]string),
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
	case "nextcloud":
		conf := nextcloud.Config{}
		p := promptui.Prompt{
			Label:   "Nextcloud URL (e.g. https://example.com)",
			Default: "",
		}
		conf.URL, err = p.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}
		p = promptui.Prompt{
			Label:   "Username",
			Default: "",
		}
		conf.Username, err = p.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		p = promptui.Prompt{
			Label:   "Password",
			Default: "",
			Mask:    '*',
		}
		conf.Password, err = p.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		p = promptui.Prompt{
			Label:   "Use Password protected link Shares [y/N] ",
			Default: "N",
		}
		linkShareWithPassword, err := p.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		usePass := strings.ToLower(linkShareWithPassword) == "y" || strings.ToLower(linkShareWithPassword) == "yes"

		randomPwChars := "32"
		if usePass {
			p = promptui.Prompt{
				Label:   "How many Random Password Chars",
				Default: "32",
			}
			randomPwChars, err = p.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}
		}

		config.ProviderSettings["username"] = conf.Username
		config.ProviderSettings["password"] = conf.Password
		config.ProviderSettings["url"] = conf.URL
		config.ProviderSettings["linkShareWithPassword"] = cast.ToString(usePass)
		config.ProviderSettings["randomPasswordChars"] = cast.ToString(randomPwChars)

	case "seafile":
		conf := seafile.Config{}
		urlPrompt := promptui.Prompt{
			Label:   "Seafile URL (e.g. https://seacloud.cc)",
			Default: "",
		}
		conf.URL, err = urlPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}
		userPrompt := promptui.Prompt{
			Label:   "Username",
			Default: "",
		}
		conf.Username, err = userPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		passwordPrompt := promptui.Prompt{
			Label:   "Password",
			Default: "",
			Mask:    '*',
		}
		conf.Password, err = passwordPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		twoFactorPrompt := promptui.Prompt{
			Label: "Is two factor auth enabled [y/N] ?",
		}
		twoFactorEnabled, err := twoFactorPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}
		if strings.ToLower(twoFactorEnabled) == "y" {
			conf.TwoFactorEnabled = true
			otpPrompt := promptui.Prompt{
				Label:   "OTP Token",
				Default: "",
			}
			conf.OTP, err = otpPrompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}
		}

		token, err := conf.GetToken()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		conf.CreateLibrary(token)
		config.ProviderSettings["token"] = token
		config.ProviderSettings["url"] = conf.URL
		config.ProviderSettings["repoid"] = conf.RepoID

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
	case "opendrive":
		var user, password string
		userPrompt := promptui.Prompt{
			Label:   "Username",
			Default: "",
		}
		user, err = userPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}

		passwordPrompt := promptui.Prompt{
			Label:   "Password",
			Default: "",
			Mask:    '*',
		}
		password, err = passwordPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}
		config.ProviderSettings["user"] = user
		config.ProviderSettings["pass"] = password
	}
	u, settings := urlshortener.Questions()
	config.URLShortenerProvider = u
	config.URLShortenerSettings = settings
	fmt.Println("write config...")
	return config.Write()
}
