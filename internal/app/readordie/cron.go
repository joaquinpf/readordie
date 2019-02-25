package readordie

import (
	"github.com/asdine/storm"
	"github.com/jasonlvhit/gocron"
	"github.com/joaquinpf/readordie/internal/pkg/readordie"
	log "github.com/sirupsen/logrus"
	"time"
)

type cronEnv struct {
	db *storm.DB
}

// NewCronEnv creates a new environment for cron jobs
func NewCronEnv(stormdb *storm.DB) cronEnv {
	return cronEnv{db: stormdb}
}

// Start initializes all cron jobs
func (env cronEnv) Start(everyMinutes uint64) {
	gocron.Every(everyMinutes).Minutes().Do(env.updateMangas)
	gocron.Start()
	gocron.RunAll()
}

func (env cronEnv) updateMangas() {
	log.Info("Checking for new manga chapters")
	var mangas []readordie.Manga
	err := env.db.AllByIndex("Name", &mangas)
	if err == storm.ErrNotFound {
		log.Info("No mangas to process")
		return
	}
	if err != nil {
		log.Error("Unable to load Manga list from DB")
		return
	}

	for _, manga := range mangas {
		log.Infof("Checking for updates for %v", manga)
		chapters, err := manga.ListChapters()
		if err != nil {
			log.Errorf("Unable to load chapters for %v", manga)
			continue
		}
		mangaRepo := env.db.From(manga.ID)
		for _, chapter := range chapters {
			var ec readordie.Chapter
			err := mangaRepo.One("ID", chapter.ID, &ec)
			if err == nil {
				continue
			}

			log.Infof("Downloading %v", chapter)
			err = readordie.DownloadAndZip(manga, chapter, manga.Folder)
			if err != nil {
				log.Errorf("Unable to download %v", chapter)
				continue
			}

			chapter.DownloadedDate = time.Now()
			err = mangaRepo.Save(&chapter)
			if err != nil {
				log.Errorf("Unable to store newly downloaded chapter in DB for %v", chapter)
				continue
			}
		}
	}
	log.Infof("Finished checking for new manga chapters")
}
