package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hirosato/wcs/env"
)

func AddCorsHeader(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", env.GetFrontUrl())
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Headers", "content-type")
}
