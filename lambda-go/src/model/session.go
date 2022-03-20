package model

import "github.com/gomodule/oauth1/oauth"

type Session struct {
	SessionId  string
	UserId     string
	TempToken  string
	TempSecret string
	Token      string
	Secret     string
}

func (s Session) IsValid() bool {
	return s.Token != "" && s.Secret != ""
}

func (s Session) IsLoggedIn() bool {
	return s.Token != "" && s.Secret != ""
}

func (s Session) AsOauthCredentials() (result *oauth.Credentials) {
	return &oauth.Credentials{
		Token:  s.Token,
		Secret: s.Secret,
	}
}
