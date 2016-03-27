package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
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
		// TODO, timeout
		resp, err := http.Post("https://github.com/login/oauth/access_token?client_id="+githubClient+"&client_secret="+githubSecret+"&code="+url.QueryEscape(req.GithubCode), "", nil)
		if err != nil {
			serverErr(w, "Error making github req: "+err.Error())
			return
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			serverErr(w, "Bad github resp: "+err.Error())
			return
		}
		vals, err := url.ParseQuery(string(respBody))
		if err != nil {
			serverErr(w, err.Error())
			return
		}
		accessToken := vals.Get("access_token")
		if accessToken == "" {
			serverErr(w, "invalid github response; empty")
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"AccessToken":"` + vals.Get("access_token") + `"}`))
		return
	}
}
