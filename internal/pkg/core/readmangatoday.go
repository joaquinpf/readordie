package core

import (
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ReadMangaToday provider implementation
type ReadMangaToday struct{}

// GetMangaLink satisfies MangaProvider interface
func (rmt ReadMangaToday) GetMangaLink(name string) (string, error) {
	urlName := strings.ToLower(strings.Replace(name, " ", "-", -1))
	mangaURL := fmt.Sprintf("%v/%v", "https://www.readmng.com", urlName)
	res, err := http.Head(mangaURL)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", errors.New("manga not found in ReadMangaToday")
	}

	return mangaURL, nil
}

// ListChapters satisfies MangaProvider interface
func (rmt ReadMangaToday) ListChapters(manga Manga) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	resp, err := soup.Get(manga.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	links := doc.Find("ul", "class", "chp_lst").FindAll("li")
	for _, row := range links {
		info := row.Find("span", "class", "val")
		if info.Error != nil {
			return nil, info.Error
		}
		infoParts := strings.Split(info.Text(), "-")
		chapterText := strings.TrimSpace(infoParts[len(infoParts)-1])
		var chapterNumberMajor uint64
		var chapterNumberMinor uint64
		var version uint64

		if strings.Contains(chapterText, "v") {
			chapterDetails := strings.Split(chapterText, "v")
			chapterNumberMajor, err = strconv.ParseUint(chapterDetails[0], 10, 16)
			if err != nil {
				return nil, err
			}
			version, err = strconv.ParseUint(chapterDetails[1], 10, 8)
			if err != nil {
				return nil, err
			}
		} else if strings.Contains(chapterText, ".") {
			chapterDetails := strings.Split(chapterText, ".")
			chapterNumberMajor, err = strconv.ParseUint(chapterDetails[0], 10, 16)
			if err != nil {
				return nil, err
			}
			chapterNumberMinor, err = strconv.ParseUint(chapterDetails[1], 10, 8)
			if err != nil {
				return nil, err
			}
		} else {
			chapterNumberMajor, err = strconv.ParseUint(chapterText, 10, 16)
			if err != nil {
				return nil, err
			}
		}

		link := row.Find("a").Attrs()["href"]
		mc := NewChapter(manga.ID, manga.Provider, uint16(chapterNumberMajor), uint8(chapterNumberMinor), uint8(version), link, time.Time{})
		chapters = append(chapters, mc)
	}
	return chapters, nil
}

// ListPages satisfies MangaProvider interface
func (rmt ReadMangaToday) ListPages(chapter Chapter) ([]Page, error) {

	pages := make([]Page, 0)
	resp, err := soup.Get(chapter.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("select", "name", "category_type")
	if len(links) < 2 {
		return nil, errors.New("Can't parse page count")
	}
	links = links[1].FindAll("option")

	for i := 0; i < len(links); i++ {
		img, err := rmt.getImageLink(links[i].Attrs()["value"])
		if err != nil {
			return nil, err
		}
		pages = append(pages, Page{Link: img, Number: uint8(i + 1)})
	}
	return pages, nil
}

func (rmt ReadMangaToday) getImageLink(pageLink string) (string, error) {
	resp, err := soup.Get(pageLink)
	if err != nil {
		return "", err
	}
	doc := soup.HTMLParse(resp)
	link := doc.Find("img", "id", "chapter_img")
	return link.Attrs()["src"], nil
}
