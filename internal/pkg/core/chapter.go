package core

import (
	"fmt"
	"time"
)

// Chapter model
type Chapter struct {
	ID             string `storm:"id"`
	MangaID        string `storm:"index"`
	Major          uint16 `storm:"index"`
	Minor          uint8  `storm:"index"`
	Version        uint8
	Provider       string
	Link           string
	DownloadedDate time.Time
}

// NewChapter creates a new Chapter instance
func NewChapter(mangaID string, provider string, major uint16, minor uint8, version uint8, link string, downloadedDate time.Time) Chapter {
	return Chapter{
		ID:             NewIDFromData([]int{int(major), int(minor)}),
		MangaID:        mangaID,
		Provider:       provider,
		Major:          major,
		Minor:          minor,
		Version:        version,
		Link:           link,
		DownloadedDate: downloadedDate}
}

// ListPages lists pages for this chapter
func (ch Chapter) ListPages() ([]Page, error) {
	provider, err := ProviderFromString(ch.Provider)
	if err != nil {
		return nil, err
	}
	return provider.ListPages(ch)
}

func (ch Chapter) String() string {
	if ch.Minor == 0 {
		return fmt.Sprintf("%v", ch.Major)
	}
	return fmt.Sprintf("%v.%v", ch.Major, ch.Minor)
}
