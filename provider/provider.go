package provider

import (
	"os"
)

// Provider Interface...
type Provider interface {
	Upload(file *os.File, path string) (string, error)
	GetLink(string) (string, error)
}
