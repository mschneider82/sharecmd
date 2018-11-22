package main

import (
	"os"
)

// Provider Interface...
type Provider interface {
	Upload(file *os.File, path string) error
	GetLink(filepath string) string
}
