package clipboard

import (
	"fmt"
	"log"

	"github.com/atotto/clipboard"
)

// ToClip copys output to clipboard
func ToClip(output string) {
	if err := clipboard.WriteAll(output); err != nil {
		log.Fatalf("Can't copy link to clipboard: %s", err.Error())
	}
	fmt.Println("URL copied to clipboard!")
}
