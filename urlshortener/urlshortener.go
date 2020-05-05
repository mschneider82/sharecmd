package urlshortener

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"schneider.vip/share/urlshortener/biturl"
)

// URLShortener Interface
type URLShortener interface {
	GetName() string
	SetupQuestions() map[string]string
	ShortURL() (string, error)
}

var providers = []string{"none", "biturl"}

// Questions ask for url shortener provider or none
func Questions() (string, map[string]string) {

	prompt := promptui.Select{
		Label: "Select an URL Shortener",
		Items: providers,
	}

	_, provider, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", make(map[string]string)
	}

	switch provider {
	case "biturl":
		b := biturl.New("")
		m := b.SetupQuestions()
		return b.GetName(), m
	default:
		return "", make(map[string]string)
	}
}
