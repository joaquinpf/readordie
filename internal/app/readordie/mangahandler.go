package readordie

import (
	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
	"github.com/joaquinpf/readordie/internal/pkg/readordie"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type mangaRequest struct {
	Name     string
	Provider string
	Folder   string
}

func (senv serverEnv) postManga(c *gin.Context) {
	var request mangaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var existing readordie.Manga
	err := senv.db.One("Name", request.Name, &existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Manga already registered"})
		return
	}

	manga, err := readordie.NewManga(request.Name, request.Folder, request.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senv.loadStoredFiles(*manga)

	err = senv.db.Save(manga)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manga)
}

func (senv serverEnv) putManga(c *gin.Context) {
	var request mangaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var em readordie.Manga
	err := senv.db.One("ID", c.Param("mid"), &em)
	if err == storm.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manga doesn't exist"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	manga, err := readordie.NewManga(em.Name, em.Folder, request.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	manga.ID = em.ID

	senv.loadStoredFiles(*manga)

	err = senv.db.Save(manga)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manga)
}

func (senv serverEnv) listManga(c *gin.Context) {
	var mangas []readordie.Manga
	err := senv.db.AllByIndex("Name", &mangas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"mangas": mangas,
	})
}

func (senv serverEnv) getManga(c *gin.Context) {
	var em readordie.Manga
	err := senv.db.One("ID", c.Param("mid"), &em)
	if err == storm.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manga Not Found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, em)
}

func (senv serverEnv) deleteManga(c *gin.Context) {
	var manga readordie.Manga
	manga.ID = c.Param("mid")
	err := senv.db.DeleteStruct(&manga)
	if err == storm.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manga Not Found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (senv serverEnv) listMangaChapters(c *gin.Context) {
	var em readordie.Manga
	err := senv.db.One("ID", c.Param("mid"), &em)
	if err == storm.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manga Not Found"})
		return
	}

	mangaRepo := senv.db.From(c.Param("mid"))
	chapters := make([]readordie.Chapter, 0)
	err = mangaRepo.AllByIndex("Major", &chapters)
	if err != nil && err != storm.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chapters": chapters})
}

func (senv serverEnv) rescanManga(c *gin.Context) {
	var mangas []readordie.Manga
	err := senv.db.AllByIndex("Name", &mangas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, manga := range mangas {
		senv.db.Drop(manga.ID)
		senv.loadStoredFiles(manga)
	}

	c.Status(http.StatusOK)
}

func (senv serverEnv) loadStoredFiles(manga readordie.Manga) {
	files, err := ioutil.ReadDir(manga.Folder)
	if err != nil {
		return
	}
	mangaRepo := senv.db.From(manga.ID)
	for _, f := range files {
		var re = regexp.MustCompile(`(?m)(?P<name>.*) - c(?P<major>[^.]+)(?:\.(?P<minor>.*))?\.(zip|rar|cbz|cbr)`)
		matches := re.FindStringSubmatch(f.Name())
		if matches == nil {
			continue
		}
		var minor uint64
		major, err := strconv.ParseUint(matches[2], 10, 16)
		if err != nil {
			continue
		}
		if matches[3] != "" {
			minor, err = strconv.ParseUint(matches[3], 10, 8)
			if err != nil {
				continue
			}
		}

		ch := readordie.NewChapter(manga.ID, manga.Provider, uint16(major), uint8(minor), 0, "", time.Now())
		err = mangaRepo.Save(&ch)
		if err != nil {
			continue
		}

		log.Infof("Loaded %v from disk for %v", ch, manga)
	}
}
