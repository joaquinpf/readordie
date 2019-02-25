package readordie

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MangaHub provider implementation
type MangaHub struct{}

// GetMangaLink satisfies MangaProvider interface
func (mh MangaHub) GetMangaLink(name string) (string, error) {
	urlName := mh.sanitize(name)
	mangaURL := fmt.Sprintf("%v/manga/%v", "http://mangahub.io", urlName)
	res, err := http.Head(mangaURL)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", errors.New("manga not found in MangaHub")
	}

	return mangaURL, nil
}

// ListChapters satisfies MangaProvider interface
func (mh MangaHub) ListChapters(manga Manga) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	resp, err := soup.Get(manga.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("a", "class", "_2U6DJ")
	for _, row := range links {
		link := row.Attrs()["href"]
		linkParts := strings.Split(link, "-")
		chapterText := linkParts[len(linkParts)-1]

		var chapterNumberMajor uint64
		var chapterNumberMinor uint64
		if strings.Contains(chapterText, ".") {
			chapterDetails := strings.Split(chapterText, ".")
			chapterNumberMajor, err = strconv.ParseUint(chapterDetails[0], 10, 16)
			if err != nil {
				log.Warnf("MangaHub: Error while parsing chapter string %v", chapterText)
				continue
			}
			chapterNumberMinor, err = strconv.ParseUint(chapterDetails[1], 10, 8)
			if err != nil {
				log.Warnf("MangaHub: Error while parsing chapter string %v", chapterText)
				continue
			}
		} else {
			chapterNumberMajor, err = strconv.ParseUint(chapterText, 10, 16)
			if err != nil {
				log.Warnf("MangaHub: Error while parsing chapter string %v", chapterText)
				continue
			}
		}

		mc := NewChapter(manga.ID, manga.Provider, uint16(chapterNumberMajor), uint8(chapterNumberMinor), 0, link, time.Time{})
		chapters = append(chapters, mc)
	}
	return chapters, nil
}

// ListPages satisfies MangaProvider interface
func (mh MangaHub) ListPages(chapter Chapter) ([]Page, error) {
	pages := make([]Page, 0)
	cQuery := fmt.Sprintf("%v", chapter.Major)
	if chapter.Minor != 0 {
		cQuery = fmt.Sprintf("%v.%v", chapter.Major, chapter.Minor)
	}

	linkParts := strings.Split(chapter.Link, "/")
	slug := linkParts[len(linkParts)-2]

	var jsonStr = fmt.Sprintf(`query {chapter(slug:"%v",number:%v){pages}}`, slug, cQuery)
	client := graphql.NewClient("https://api2.mangahub.io/graphql")

	req := graphql.NewRequest(jsonStr)

	var responseData map[string]interface{}
	err := client.Run(context.Background(), req, &responseData)
	if err != nil {
		return nil, err
	}

	responsePages := responseData["chapter"].(map[string]interface{})["pages"].(string)
	var result map[uint8]string
	err = json.Unmarshal([]byte(responsePages), &result)
	if err != nil {
		return nil, err
	}

	for k, v := range result {
		link := fmt.Sprintf("%v/%v", "https://cdn.mangahub.io/file/imghub", v)
		pages = append(pages, Page{Link: link, Number: k})
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Number < pages[j].Number
	})

	return pages, nil
}

func (mh MangaHub) sanitize(name string) string {
	sanitized := strings.Replace(name, " ", "-", -1)
	sanitized = strings.Replace(sanitized, ".", "-", -1)
	sanitized = strings.Replace(sanitized, "!", "", -1)
	return strings.ToLower(sanitized)
}
