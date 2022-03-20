package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hirosato/wcs/db"
	"github.com/hirosato/wcs/model"
)

func ServePigmentSearch(c *gin.Context) {
	cat := c.Query("cat")
	q := c.Query("q")
	if cat == "" || q == "" {
		c.JSON(400, gin.H{
			"message": "bad request",
		})
		return
	}
	icat, err := strconv.Atoi(cat)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "bad request",
		})
		return
	}

	equipments, err := db.GetPigment(model.JA, int32(icat), q)
	if err != nil {
		c.JSON(404, gin.H{
			"message": "error: " + err.Error(),
		})
		return
	} else {
		c.JSON(200, gin.H{
			"results": equipments,
		})
	}
}
