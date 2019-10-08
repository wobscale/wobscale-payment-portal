package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
)

type GithubLoginRequest struct {
	GithubCode string
}

func GithubLogin(githubSecret, githubClient string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		var req GithubLoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		r.Body.Close()

		if err != nil {
			userErr(w, "Malformed request")
			return
		}

		if req.GithubCode == "" {
			userErr(w, "GithubCode required")
			return
		}

		// The go-github library doesn't have an accessToken api, just do it by hand
		timeoutCtx, cancel := context.WithTimeout(context.TODO(), 4*time.Second)
		defer cancel()
		httpReq, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token?client_id="+githubClient+"&client_secret="+githubSecret+"&code="+url.QueryEscape(req.GithubCode), nil)
		if err != nil {
			serverErrf(w, "Error constructing github req: %s", err)
			return
		}
		httpReq.Header.Set("Accept", "application/json")
		httpReq = httpReq.WithContext(timeoutCtx)
		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			serverErrf(w, "Error making github req: %s", err)
			return
		}
		if resp.StatusCode >= 400 {
			serverErrf(w, "Github returned an error status: %d", resp.StatusCode)
		}

		logrus.Debugf("github response: %v", resp)
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			serverErrf(w, "Bad github resp: %s", err)
			return
		}
		var githubResp githubOauthResponse
		if err := json.Unmarshal(respBody, &githubResp); err != nil {
			logrus.Warnf("expected json, got %s", respBody)
			serverErr(w, "invalid github oauth response; not valid json")
			return
		}
		if githubResp.AccessToken == "" {
			logrus.Debugf("empty access token: %s", respBody)
			serverErr(w, "invalid github response; no auth token")
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(struct {
			AccessToken string
		}{
			githubResp.AccessToken,
		})
		return
	}
}

type githubOauthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}
