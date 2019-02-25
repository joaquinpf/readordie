package readordie

import (
	"errors"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/anaskhan96/soup"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MangaDex provider implementation
type MangaDex struct{}

// GetMangaLink satisfies MangaProvider interface
func (md MangaDex) GetMangaLink(name string) (string, error) {
	searchURL := fmt.Sprintf("%v/quick_search/%v", "https://mangadex.org", url.PathEscape(name))
	resp, err := soup.Get(searchURL)
	if err != nil {
		return "", errors.New("manga not found in MangaDex")
	}
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("a", "class", "manga_title")
	minimumDistance := 1000
	var mangaURL string
	for _, row := range links {
		rowTitle := row.Attrs()["title"]
		link := row.Attrs()["href"]
		distance := levenshtein.ComputeDistance(name, rowTitle)
		if distance < minimumDistance {
			mangaURL = "https://mangadex.org" + link
			minimumDistance = distance
		}
	}

	return mangaURL, nil
}

// ListChapters satisfies MangaProvider interface
func (md MangaDex) ListChapters(manga Manga) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	soup.Cookie("mangadex_filter_langs", "1")
	totalPages, err := md.getMaxPagination(manga)
	if err != nil {
		return nil, err
	}

	for page := 1; page <= totalPages; page++ {
		resp, err := soup.Get(fmt.Sprintf("%v/%v", manga.Link, page))
		if err != nil {
			return nil, err
		}
		doc := soup.HTMLParse(resp)
		links := doc.FindAll("div", "class", "chapter-row")
		for _, row := range links {
			a := row.Find("a", "class", "text-truncate")
			if a.Error != nil {
				continue
			}

			link := "https://mangadex.org" + a.Attrs()["href"]
			chapterText := row.Attrs()["data-chapter"]

			var chapterNumberMajor uint64
			var chapterNumberMinor uint64
			if strings.Contains(chapterText, ".") {
				chapterDetails := strings.Split(chapterText, ".")
				chapterNumberMajor, err = strconv.ParseUint(chapterDetails[0], 10, 16)
				if err != nil {
					log.Warnf("MangaDex: Error while parsing chapter string %v", chapterText)
					continue
				}
				chapterNumberMinor, err = strconv.ParseUint(chapterDetails[1], 10, 8)
				if err != nil {
					log.Warnf("MangaDex: Error while parsing chapter string %v", chapterText)
					continue
				}
			} else {
				chapterNumberMajor, err = strconv.ParseUint(chapterText, 10, 16)
				if err != nil {
					log.Warnf("MangaDex: Error while parsing chapter string %v", chapterText)
					continue
				}
			}

			mc := NewChapter(manga.ID, manga.Provider, uint16(chapterNumberMajor), uint8(chapterNumberMinor), 0, link, time.Time{})
			chapters = append(chapters, mc)
		}
	}
	return chapters, nil
}

func (md MangaDex) getMaxPagination(manga Manga) (int, error) {
	resp, err := soup.Get(manga.Link)
	if err != nil {
		return -1, err
	}
	doc := soup.HTMLParse(resp)
	pages := doc.FindAll("a", "class", "page-link")
	lastPage := pages[len(pages)-1].Attrs()["href"]
	pieces := strings.Split(lastPage, "/")
	return strconv.Atoi(pieces[len(pieces)-2])
}

// ListPages satisfies MangaProvider interface
func (md MangaDex) ListPages(chapter Chapter) ([]Page, error) {
	pages := make([]Page, 0)
	resp, err := soup.Get(chapter.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	totalPageElement := doc.Find("span", "class", "total-pages")
	if totalPageElement.Error != nil {
		return nil, errors.New("Can't parse page count")
	}
	totalPages, err := strconv.Atoi(totalPageElement.Text())
	if err != nil {
		return nil, errors.New("Can't parse page count")
	}

	img := doc.Find("div", "class", "reader-image-wrapper").Find("img")
	pages = append(pages, Page{Link: img.Attrs()["src"], Number: 1})

	for page := 2; page <= totalPages; page++ {
		pageLink := fmt.Sprintf("%v/%v", chapter.Link, page)
		link, err := md.getImageLink(pageLink)
		if err != nil {
			return nil, err
		}
		pages = append(pages, Page{Link: link, Number: uint8(page)})
	}
	return pages, nil
}

func (md MangaDex) getImageLink(pageLink string) (string, error) {
	resp, err := soup.Get(pageLink)
	if err != nil {
		return "", err
	}
	doc := soup.HTMLParse(resp)
	img := doc.Find("div", "class", "reader-image-wrapper").Find("img")
	if img.Error != nil {
		return "", img.Error
	}

	return img.Attrs()["src"], nil
}
