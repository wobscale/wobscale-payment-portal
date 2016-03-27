package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

type CustomerMetadata string

const (
	GithubUsernameMetadata CustomerMetadata = "GithubUsername"
	GithubUserIDMetadata   CustomerMetadata = "GithubUserID"
)

type SubPlan struct {
	Name string
	Cost uint64
	Num  uint64
}

type SubscriptionRequest struct {
	Nickname          string
	GithubAccessToken string
	Email             string
	StripeToken       string
}

type apiErr struct {
	Error string
}

func NewSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var req SubscriptionRequest
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

	if err != nil {
		userErr("Malformed request")
		return
	}

	if req.GithubAccessToken == "" {
		userErr("Github auth is required")
		return
	}

	if req.Email == "" {
		userErr("Email is required")
		return
	}

	if req.StripeToken == "" {
		userErr("Stripe setup (CC etc) is required")
		return
	}

	// Validate githurb stuff correctly now
	githubauth := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: req.GithubAccessToken})
	githurb := github.NewClient(oauth2.NewClient(oauth2.NoContext, githubauth))
	authedUser, _, err := githurb.Users.Get("")
	if err != nil {
		logrus.Info("Error with user's github auth key: ", err)
		// Could be a server-err from github, TODO, clasify github errors as 400/500 and mimic
		userErr("Github auth didn't appear to work")
		return
	}
	if authedUser.Login == nil {
		logrus.Errorf("Nil github user name: %v", authedUser)
		serverErr("Github api issue")
		return
	}
	if authedUser.ID == nil {
		logrus.Errorf("Nil userid: %v", authedUser)
		serverErr("Github api issue")
		return
	}
	githubUid := strconv.Itoa(*authedUser.ID)

	// valid, let's create this customer

	customerParams := &stripe.CustomerParams{
		Email: req.Email,
		Desc:  req.Nickname,
	}
	customerParams.IdempotencyKey = githubUid
	customerParams.Meta = map[string]string{
		string(GithubUserIDMetadata):   githubUid,
		string(GithubUsernameMetadata): *authedUser.Login,
	}
	err = customerParams.SetSource(req.StripeToken)
	// TODO, check if error is 400 or 500
	if err != nil {
		logrus.Infof("Invalid stripe token: %v", err)
		userErr("Invalid stripe token")
		return
	}

	createdCustomer, err := customer.New(customerParams)
	if err != nil {
		logrus.Warnf("Error creating customer: %v", err)
		serverErr("Error with stripe api for creating customer")
		return
	}
	logrus.Infof("Customer created: %v, %v", req.Email, createdCustomer.ID)

	// TODO, email admin@ or something?
	// TODO, optimize by returning customer-id here, adding a getter that takes it

	w.WriteHeader(200)
	w.Write([]byte(`{"ok": 1}`))
}
