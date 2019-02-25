package readordie

import (
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MangaSee provider implementation
type MangaSee struct{}

// GetMangaLink satisfies MangaProvider interface
func (ms MangaSee) GetMangaLink(name string) (string, error) {
	urlName := ms.sanitize(name)
	mangaURL := fmt.Sprintf("%v/manga/%v", "http://mangaseeonline.us", urlName)
	res, err := http.Head(mangaURL)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", errors.New("manga not found in MangaSee")
	}

	return mangaURL, nil
}

func (ms MangaSee) sanitize(name string) string {
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}-]+`)
	urlName := re_inside_whtsp.ReplaceAllString(name, "-")
	urlName = strings.Replace(urlName, ":", "-", -1)
	return urlName
}

// ListChapters satisfies MangaProvider interface
func (ms MangaSee) ListChapters(manga Manga) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	resp, err := soup.Get(manga.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("a", "class", "list-group-item")
	for _, row := range links {
		link := fmt.Sprintf("%v%v", "http://mangaseeonline.us", row.Attrs()["href"])
		chapterText := row.Attrs()["chapter"]

		var chapterNumberMajor uint64
		var chapterNumberMinor uint64
		if strings.Contains(chapterText, ".") {
			chapterDetails := strings.Split(chapterText, ".")
			chapterNumberMajor, err = strconv.ParseUint(chapterDetails[0], 10, 16)
			if err != nil {
				log.Warnf("MangaSee: Error while parsing chapter string %v", chapterText)
				continue
			}
			chapterNumberMinor, err = strconv.ParseUint(chapterDetails[1], 10, 8)
			if err != nil {
				log.Warnf("MangaSee: Error while parsing chapter string %v", chapterText)
				continue
			}
		} else {
			chapterNumberMajor, err = strconv.ParseUint(chapterText, 10, 16)
			if err != nil {
				log.Warnf("MangaSee: Error while parsing chapter string %v", chapterText)
				continue
			}
		}

		mc := NewChapter(manga.ID, manga.Provider, uint16(chapterNumberMajor), uint8(chapterNumberMinor), 0, link, time.Time{})
		chapters = append(chapters, mc)
	}
	return chapters, nil
}

// ListPages satisfies MangaProvider interface
func (ms MangaSee) ListPages(chapter Chapter) ([]Page, error) {
	pages := make([]Page, 0)
	resp, err := soup.Get(chapter.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("select", "class", "PageSelect")
	if len(links) < 1 {
		return nil, errors.New("Can't parse page count")
	}
	links = links[0].FindAll("option")

	linkParts := strings.Split(chapter.Link, "-")

	for i := 1; i <= len(links); i++ {
		linkParts[len(linkParts)-1] = fmt.Sprintf("%v.html", i)
		pageLink := strings.Join(linkParts[:], "-")

		img, err := ms.getImageLink(pageLink)
		if err != nil {
			return nil, err
		}
		pages = append(pages, Page{Link: img, Number: uint8(i)})
	}
	return pages, nil
}

func (ms MangaSee) getImageLink(pageLink string) (string, error) {
	resp, err := soup.Get(pageLink)
	if err != nil {
		return "", err
	}
	doc := soup.HTMLParse(resp)
	link := doc.Find("div", "class", "image-container-manga").Find("img")
	return link.Attrs()["src"], nil
}
