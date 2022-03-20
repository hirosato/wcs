package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hirosato/wcs/db"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/model"
	"github.com/hirosato/wcs/session"
	"github.com/hirosato/wcs/twitter"
)

func Login(c *gin.Context) {
	alreadySignedInOnTwitter, redirectUrl, _ := twitter.ServeSignin(c.Writer, c.Request)
	if alreadySignedInOnTwitter {
		sess := session.GetSession(c.Request)
		twitterUser, _ := twitter.GetTwitterUserInfo(c.Writer, c.Request)
		if twitterUser == nil {
			c.JSON(404, gin.H{
				"message": "404",
			})
		} else {
			user := twitterUser.AsUser()
			db.PutUser(user)
			sess.UserId = user.UserId
			session.SetSession(c.Writer, c.Request, sess)

			c.JSON(200, gin.H{
				"message": "pong " + user.UserId + " " + user.DisplayName + " " + user.AvatarURL + " ",
			})
		}
	} else {
		http.Redirect(c.Writer, c.Request, redirectUrl, http.StatusFound)
	}
}

func Callback(c *gin.Context) {
	twitter.ServeOAuthCallback(c.Writer, c.Request)
	twitterUser, _ := twitter.GetTwitterUserInfo(c.Writer, c.Request)
	if twitterUser == nil {
		c.JSON(404, gin.H{
			"message": "404",
		})
	} else {
		sess := session.GetSession(c.Request)
		user := twitterUser.AsUser()
		db.PutUser(user)
		sess.UserId = user.UserId
		session.SetSession(c.Writer, c.Request, sess)

		http.Redirect(c.Writer, c.Request, env.GetFrontUrl(), http.StatusFound)
	}

}

func ServeGetUser(c *gin.Context) {
	sess := session.GetSession(c.Request)
	if !sess.IsLoggedIn() {
		c.JSON(400, gin.H{
			"isLoggedIn": sess.IsLoggedIn(),
		})
		return
	}
	if sess.UserId == "" {
		c.JSON(400, gin.H{
			"isLoggedIn": false,
		})
		log.Printf("session.UserId is not set. session: %s", sess.SessionId)
		db.PutSession(model.Session{SessionId: sess.SessionId})
		return
	}
	user, err := db.GetUser(sess.UserId)
	if err != nil {
		c.JSON(500, gin.H{
			"isLoggedIn": sess.IsLoggedIn(),
		})
		db.PutSession(model.Session{SessionId: sess.SessionId})
		return
	}
	c.JSON(200, user)
}

func GetUser(r *http.Request) (user model.User, err error) {
	sess := session.GetSession(r)
	if !sess.IsLoggedIn() {
		return model.User{}, errors.New("not logged in")
	}
	if sess.UserId == "" {
		return model.User{}, errors.New("no user id")
	}
	user, err = db.GetUser(sess.UserId)
	if err != nil {
		return user, err
	}
	return user, nil
}
