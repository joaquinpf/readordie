package readordie

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (senv serverEnv) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
