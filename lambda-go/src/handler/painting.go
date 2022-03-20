package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hirosato/wcs/db"
	"github.com/hirosato/wcs/domain"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/file"
	"github.com/hirosato/wcs/model"
	"github.com/hirosato/wcs/util"
)

var filerepo domain.LocalFileRepository = file.NewLocalFileRepository()
var s3repo domain.S3Repository = file.NewS3RepositoryImpl()

func parseBody(c *gin.Context) (*model.Painting, error) {
	var user model.User
	var err error
	if user, err = GetUser(c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return &model.Painting{}, err
	}
	var painting model.Painting
	if err := c.ShouldBindJSON(&painting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return &model.Painting{}, err
	}
	painting.UserId = user.UserId
	painting.Date, painting.Timestamp = util.GetDateAndTimestamp()
	if env.IsLocal {
		painting.Date = "20210811"
		painting.Timestamp = "20210811150726359"
	}
	return &painting, nil
}

func parseImageBody(c *gin.Context) (*model.PaintingImage, error) {
	id := c.Param("id")
	timestamp := c.Param("timestamp")
	var err error
	if _, err = GetUser(c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return &model.PaintingImage{}, err
	}
	var paintingImages model.PaintingImage
	if err := c.ShouldBindJSON(&paintingImages); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return &model.PaintingImage{}, err
	}
	if id != paintingImages.UserId || timestamp != paintingImages.Timestamp {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stop it. we see you."})
	}
	return &paintingImages, nil
}

func Submit(c *gin.Context) {
	log.Printf("EVENT: Submit start")
	painting, err := parseBody(c)

	if err != nil {
		return
	}

	log.Printf("EVENT: Submitting %s", painting.GetId())
	err = db.PutPainting(painting)
	log.Printf("EVENT: Submitting %s, dynamo done", painting.GetId())
	db.PutEsPainting(painting)
	log.Printf("EVENT: Submitting %s, es done", painting.GetId())

	if err != nil {
		c.JSON(500, painting)
	} else {
		c.JSON(200, painting)
	}

	log.Printf("EVENT: Submit end")
}

func ServeSubmitPreflight(c *gin.Context) {
	c.Header("Access-Control-Allow-Headers", "content-type")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH")
	c.Status(200)
}

func uploadFile(userId string, timestamp string, base64image string, imageKind model.ImageKind) error {
	filename, err := filerepo.Add(userId, timestamp, base64image, imageKind)
	if err != nil {
		return err
	}
	err = s3repo.Add(filename, userId, timestamp, imageKind)
	if err != nil {
		filerepo.Remove(filename)
		return err
	}
	filerepo.Remove(filename)
	return nil
}

func ServePaintingList(c *gin.Context) {
	c.JSON(200, db.ListWaterColorSite(0))
}

//wcs/:id
func ServeUserPainting(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "/wcs/:id",
	})
}

//PATCH /wcs
func PatchPaintingImage(c *gin.Context) {
	log.Printf("EVENT: patch start")
	paintingImages, err := parseImageBody(c)
	if err != nil {
		return
	}
	if paintingImages.ImageCover != "" {
		err = uploadFile(paintingImages.UserId, paintingImages.Timestamp, paintingImages.ImageCover, model.ImageCover)
	} else if paintingImages.Image1 != "" {
		err = uploadFile(paintingImages.UserId, paintingImages.Timestamp, paintingImages.Image1, model.Image1)
	} else if paintingImages.Image2 != "" {
		err = uploadFile(paintingImages.UserId, paintingImages.Timestamp, paintingImages.Image2, model.Image2)
	} else if paintingImages.Image3 != "" {
		err = uploadFile(paintingImages.UserId, paintingImages.Timestamp, paintingImages.Image3, model.Image3)
	} else if paintingImages.Image4 != "" {
		err = uploadFile(paintingImages.UserId, paintingImages.Timestamp, paintingImages.Image4, model.Image4)
	}
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, paintingImages)
	} else {
		c.JSON(200, paintingImages)
	}
	log.Printf("EVENT: patch end")
}

//POST /invalidate
func InvalidatePainting(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "/wcs/:id",
	})
}

//:id/:timestamp
func ServePainting(c *gin.Context) {
	id := c.Param("id")
	timestamp := c.Param("timestamp")
	var painting model.Painting
	var err error
	if env.IsLocal {
		painting, err = db.GetPainting("875895056078483456", "20210810150726359")
	} else {
		painting, err = db.GetPainting(id, timestamp)
	}
	if err != nil {
		c.JSON(404, gin.H{
			"message": "error: " + err.Error() + " id: " + id,
		})
		return
	}
	c.JSON(200, painting)
	// eq, err := db.GetPigment()
	// if err != nil {
	// 	c.JSON(404, gin.H{
	// 		"message": "error: " + err.Error() + " id: " + id,
	// 	})
	// } else {
	// 	c.JSON(200, gin.H{
	// 		"message": "pong " + painting.UserId + painting.Timestamp + " with equipment: " + eq,
	// 	})
	// }
	// database.Get(id, "1", &results)
}
