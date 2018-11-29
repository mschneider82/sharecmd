package clipboard

import (
	"log"

	"github.com/atotto/clipboard"
)

// ToClip copys output to clipboard
func ToClip(output string) {
	if err := clipboard.WriteAll(output); err != nil {
		log.Fatalf("Can't copy link to clipboard: %s", err.Error())
	}
	log.Println("URL copied to clipboard!")
}
