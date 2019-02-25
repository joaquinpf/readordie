package readordie

import (
	"errors"
)

// MangaProvider interface for external manga providers
type MangaProvider interface {
	ListChapters(manga Manga) ([]Chapter, error)
	ListPages(chapter Chapter) ([]Page, error)
	GetMangaLink(name string) (string, error)
}

// ProviderFromString provides a factory to get a MangaProvider instance from a type string
func ProviderFromString(provider string) (MangaProvider, error) {
	switch provider {
	case "ReadMangaToday":
		return ReadMangaToday{}, nil
	case "MangaStream":
		return MangaStream{}, nil
	case "MangaHub":
		return MangaHub{}, nil
	case "MangaSee":
		return MangaSee{}, nil
	default:
		return nil, errors.New("Unsupported provider")
	}
}
