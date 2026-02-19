package provider

import (
	"io"
)

// Provider Interface...
type Provider interface {
	Upload(r io.Reader, filename string, size int64) (string, error)
	GetLink(string) (string, error)
}
