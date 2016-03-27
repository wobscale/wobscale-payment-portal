package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

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

		userErr := func(err string) {
			logrus.Info("Subscription request user error: " + err)
			w.WriteHeader(400)
			out, _ := json.Marshal(apiErr{err})
			w.Write(out)
		}

		serverErr := func(err string) {
			w.WriteHeader(500)
			out, _ := json.Marshal(apiErr{err})
			w.Write(out)
		}
		if false {
			serverErr("yes")
			return
		}

		if err != nil {
			userErr("Malformed request")
			return
		}

		if req.GithubCode == "" {
			userErr("GithubCode required")
			return
		}

		// The go-github library doesn't have an accessToken api, just do it by hand
		// TODO, timeout
		resp, err := http.Post("https://github.com/login/oauth/access_token?client_id="+githubClient+"&client_secret="+githubSecret+"&code="+url.QueryEscape(req.GithubCode), "", nil)
		if err != nil {
			serverErr("Error making github req: " + err.Error())
			return
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			serverErr("Bad github resp: " + err.Error())
			return
		}
		vals, err := url.ParseQuery(string(respBody))
		if err != nil {
			serverErr(err.Error())
			return
		}
		accessToken := vals.Get("access_token")
		if accessToken == "" {
			serverErr("invalid github response; empty")
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"AccessToken":"` + vals.Get("access_token") + `"}`))
		return
	}
}
