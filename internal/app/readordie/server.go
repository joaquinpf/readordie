package readordie

import (
	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
)

type serverEnv struct {
	db *storm.DB
}

// NewServerEnv creates a new web server environment
func NewServerEnv(stormdb *storm.DB) serverEnv {
	return serverEnv{db: stormdb}
}

// Start initializes the API
func (senv serverEnv) Start() {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.GET("/status", senv.getStatus)
		v1.GET("/provider", senv.getProviders)
		v1.POST("/manga", senv.postManga)
		v1.PUT("/manga/:mid", senv.putManga)
		v1.GET("/manga", senv.listManga)
		v1.GET("/manga/:mid", senv.getManga)
		v1.DELETE("/manga/:mid", senv.deleteManga)
		v1.GET("/manga/:mid/chapter", senv.listMangaChapters)
		v1.GET("/admin/rescan", senv.rescanManga)
	}
	r.Run()
}
