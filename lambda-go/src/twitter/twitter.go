package twitter

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gomodule/oauth1/oauth"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/model"
	"github.com/hirosato/wcs/session"
)

const (
	tempCredKey  = "tempCred"
	tokenCredKey = "tokenCred"
)

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
	TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
}

func init() {
	oauthClient.Credentials.Token = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	oauthClient.Credentials.Secret = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	oauthClient.ResourceOwnerAuthorizationURI = "https://api.twitter.com/oauth/authenticate"
}

func getCallbackURL() string {
	return env.GetApiUrl() + "/twitter/callback"
}

func ServeSignin(w http.ResponseWriter, r *http.Request) (alreadySignedIn bool, redirectUrl string, err error) {
	callback := getCallbackURL()
	s := session.GetSession(r)
	if s.IsValid() {
		return true, "", nil
	}
	tempCred, err := oauthClient.RequestTemporaryCredentials(nil, callback, nil)
	if err != nil {
		http.Error(w, "Error getting temp cred, "+err.Error(), http.StatusInternalServerError)
		return false, "", err
	}
	s.TempToken = tempCred.Token
	s.TempSecret = tempCred.Secret
	if err := session.SetSession(w, r, s); err != nil {
		http.Error(w, "Error saving session , "+err.Error(), http.StatusInternalServerError)
		return false, "", err
	}
	return false, oauthClient.AuthorizationURL(tempCred, nil), nil
}

type Account struct {
	ID              string `json:"id_str"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url_https"`
}

func (account Account) AsUser() model.User {
	return model.User{
		UserId:      account.ID,
		DisplayName: account.ScreenName,
		AvatarURL:   account.ProfileImageURL,
	}
}

func ServeOAuthCallback(w http.ResponseWriter, r *http.Request) {
	s := session.GetSession(r)
	tempCred := oauth.Credentials{
		Token:  s.TempToken,
		Secret: s.TempSecret,
	}
	if tempCred.Token != r.FormValue("oauth_token") {
		http.Error(w, "Unknown oauth_token.", http.StatusInternalServerError)
		return
	}
	tokenCred, _, err := oauthClient.RequestToken(nil, &tempCred, r.FormValue("oauth_verifier"))
	if err != nil {
		http.Error(w, "Error getting request token, "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = session.SetSession(w, r, model.Session{
		SessionId:  s.SessionId,
		TempToken:  ``,
		TempSecret: ``,
		Token:      tokenCred.Token,
		Secret:     tokenCred.Secret,
	})
	if err != nil {
		http.Error(w, "Error saving session , "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetTwitterUserInfo(w http.ResponseWriter, r *http.Request) (*Account, error) {
	var user *Account = &Account{}
	s := session.GetSession(r)

	if !s.IsValid() {
		return nil, nil
	}

	resp, err := oauthClient.Get(nil, s.AsOauthCredentials(), "https://api.twitter.com/1.1/account/verify_credentials.json", url.Values{})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 500 {
		return nil, err //errors.New("Twitter is unavailable")
	}

	if resp.StatusCode >= 400 {
		return nil, err //errors.New("Twitter request is invalid")
	}

	err = json.NewDecoder(resp.Body).Decode(user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func ServeLogout(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	delete(s, tokenCredKey)
	if err := session.Save(w, r, s); err != nil {
		http.Error(w, "Error saving session , "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
