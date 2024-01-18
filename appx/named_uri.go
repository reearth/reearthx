package appx

import (
	"bytes"
	"fmt"
)

type NamedURI struct {
	Name string
	URI  string
}

func (n *NamedURI) UnmarshalText(text []byte) error {
	name, uri, found := bytes.Cut(text, []byte("="))
	if !found {
		return fmt.Errorf("invalid named uri")
	}
	n.Name = string(name)
	n.URI = string(uri)
	return nil
}
