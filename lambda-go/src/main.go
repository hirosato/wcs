package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/handler"
)

var ginLambda *ginadapter.GinLambda

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	fmt.Println("IS_LOCAL", os.Getenv("IS_LOCAL"))
	log.Printf("Gin cold start")
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/wcs", handler.AddCorsHeader, handler.ServePaintingList)
	r.GET("/wcs/:id", handler.AddCorsHeader, handler.ServeUserPainting)
	r.GET("/wcs/:id/:timestamp", handler.AddCorsHeader, handler.ServePainting)
	r.PATCH("/wcs/:id/:timestamp/images", handler.AddCorsHeader, handler.PatchPaintingImage)
	r.OPTIONS("/wcs/:id/:timestamp/images", handler.AddCorsHeader, handler.ServeSubmitPreflight)
	r.GET("/equipments", handler.AddCorsHeader, handler.ServePigmentSearch)
	r.POST("/wcs", handler.AddCorsHeader, handler.Submit)
	r.POST("/invalidate", handler.AddCorsHeader, handler.InvalidatePainting)
	r.OPTIONS("/wcs", handler.AddCorsHeader, handler.ServeSubmitPreflight)
	r.GET("/getUser", handler.AddCorsHeader, handler.ServeGetUser)
	r.GET("twitter/signin", handler.AddCorsHeader, handler.Login)
	r.GET("twitter/callback", handler.AddCorsHeader, handler.Callback)
	if env.IsLocal {
		log.Fatal(http.ListenAndServe(":8080", r))
	} else {
		ginLambda = ginadapter.New(r)
		lambda.Start(Handler)
	}

}
