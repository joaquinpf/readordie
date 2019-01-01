package readordie

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (senv serverEnv) getProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": [2]string{"ReadMangaToday", "MangaStream"},
	})
}
