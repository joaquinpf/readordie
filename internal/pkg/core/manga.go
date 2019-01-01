package core

import (
	"strings"
)

// Manga model
type Manga struct {
	ID       string `storm:"id"`
	Link     string
	Name     string `storm:"index"`
	Provider string
	Folder   string
}

// NewManga creates a new Manga instance
func NewManga(name string, folder string, provider string) (*Manga, error) {
	p, err := ProviderFromString(provider)
	if err != nil {
		return nil, err
	}

	link, err := p.GetMangaLink(name)
	if err != nil {
		return nil, err
	}

	return &Manga{ID: NewID(), Name: name, Folder: folder, Link: link, Provider: provider}, nil
}

// ListChapters lists all chapters for the current manga available in the provider
func (m Manga) ListChapters() ([]Chapter, error) {
	provider, err := ProviderFromString(m.Provider)
	if err != nil {
		return nil, err
	}

	return provider.ListChapters(m)
}

func (m Manga) String() string {
	return strings.Title(m.Name)
}
