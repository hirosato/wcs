package session

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hirosato/wcs/db"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/model"
)

var (
	mu       sync.Mutex
	sessions = make(map[string]map[string]interface{})
)

// Get returns the session data for the request client.
func Get(r *http.Request) (s map[string]interface{}) {
	if c, _ := r.Cookie("session"); c != nil && c.Value != "" {
		mu.Lock()
		s = sessions[c.Value]
		mu.Unlock()
	}
	if s == nil {
		s = make(map[string]interface{})
	}
	return s
}

func GetSession(r *http.Request) (session model.Session) {
	var err error
	c, _ := r.Cookie("session")
	// ブラウザからセッションは送られてきている
	if c != nil && c.Value != "" {
		session, err = db.GetSession(c.Value)

		// DBでは消されている。
		if err != nil {
			log.Printf("returning empty session object for %s since there is no data in db", c)
			return model.Session{}
		} else {
			return session
		}
	}
	log.Printf("returning empty session object for %s. request is from host:%s request:%s", c, r.Host, r.RequestURI)
	return model.Session{}
}

func SetSession(w http.ResponseWriter, r *http.Request, session model.Session) (err error) {
	key := ""
	if c, _ := r.Cookie("session"); c != nil {
		key = c.Value
	}
	if key == "" || session.SessionId != key {
		var buf [16]byte
		_, err = rand.Read(buf[:])
		if err != nil {
			return err
		}
		key = hex.EncodeToString(buf[:])
		session.SessionId = key
		_, err = db.PutSession(session)
		if err != nil {
			return err
		}
		domain := "watercolor.site"
		if env.IsLocal {
			domain = ""
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Path:     "/",
			Domain:   domain,
			Secure:   !env.IsLocal,
			HttpOnly: true,
			Value:    key,
			Expires:  time.Unix(5000000000, 0),
		})
		return nil
	}
	_, err = db.PutSession(session)
	if err != nil {
		return err
	}
	return nil
}

// Save saves session for the request client.
func Save(w http.ResponseWriter, r *http.Request, s map[string]interface{}) error {
	key := ""
	if c, _ := r.Cookie("session"); c != nil {
		key = c.Value
	}
	if len(s) == 0 {
		if key != "" {
			mu.Lock()
			delete(sessions, key)
			mu.Unlock()
		}
		return nil
	}
	if key == "" {
		var buf [16]byte
		_, err := rand.Read(buf[:])
		if err != nil {
			return err
		}
		key = hex.EncodeToString(buf[:])
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Path:     "/",
			Secure:   !env.IsLocal,
			HttpOnly: true,
			Value:    key,
		})
	}
	mu.Lock()
	sessions[key] = s
	mu.Unlock()
	return nil
}
