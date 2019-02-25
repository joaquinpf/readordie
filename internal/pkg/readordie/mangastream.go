package readordie

import (
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MangaStream provider implementation
type MangaStream struct{}

// GetMangaLink satisfies MangaProvider interface
func (ms MangaStream) GetMangaLink(name string) (string, error) {
	urlName := strings.Replace(name, " ", "_", -1)
	urlName = strings.Replace(urlName, ":", "_", -1)
	urlName = strings.Replace(urlName, "/", "_", -1)
	urlName = strings.ToLower(urlName)
	mangaURL := fmt.Sprintf("%v/manga/%v", "https://readms.net", urlName)
	res, err := http.Get(mangaURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", errors.New("Manga not found in MangaStream")
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	if strings.Contains(bodyString, "Page Not Found") {
		return "", errors.New("Manga not found in MangaStream")
	}

	return mangaURL, nil
}

// ListChapters satisfies MangaProvider interface
func (ms MangaStream) ListChapters(manga Manga) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	resp, err := soup.Get(manga.Link)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	div := doc.Find("div", "class", "col-sm-8")
	if div.Error != nil {
		return nil, div.Error
	}
	table := div.Find("table")
	if table.Error != nil {
		return nil, table.Error
	}
	links := table.FindAll("tr")
	for _, row := range links {
		cells := row.FindAll("td")
		if len(cells) == 0 {
			continue
		}
		info := cells[0].Find("a")
		if info.Error != nil {
			continue
		}
		infoParts := strings.Split(info.Text(), "-")
		chapterText := strings.TrimSpace(infoParts[0])
		chapterNumberMajor, err := strconv.ParseUint(chapterText, 10, 16)
		if err != nil {
			log.Warnf("MangaStream: Error while parsing chapter string %v", chapterText)
			continue
		}

		link := "https://readms.net" + info.Attrs()["href"]
		mc := NewChapter(manga.ID, manga.Provider, uint16(chapterNumberMajor), 0, 0, link, time.Time{})
		chapters = append(chapters, mc)
	}
	return chapters, nil
}

// ListPages satisfies MangaProvider interface
func (ms MangaStream) ListPages(chapter Chapter) ([]Page, error) {

	pages := make([]Page, 0)

	res, err := http.Get(chapter.Link)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	var re = regexp.MustCompile(`(?s)Last Page \((?P<pageCount>\d*)\)`)
	count, err := strconv.ParseInt(re.FindStringSubmatch(bodyString)[1], 10, 16)
	if err != nil {
		return nil, err
	}

	urlPrefix := chapter.Link[:len(chapter.Link)-2]
	var i int64
	for i = 1; i <= count; i++ {
		img, err := ms.getImageLink(fmt.Sprintf("%v/%v", urlPrefix, i))
		if err != nil {
			return nil, err
		}
		pages = append(pages, Page{Link: img, Number: uint8(i + 1)})
	}
	return pages, nil
}

func (ms MangaStream) getImageLink(pageLink string) (string, error) {
	resp, err := soup.Get(pageLink)
	if err != nil {
		return "", err
	}
	doc := soup.HTMLParse(resp)
	link := doc.Find("img", "id", "manga-page")
	return "https:" + link.Attrs()["src"], nil
}
