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

type UpdatePaymentRequest struct {
	GithubAccessToken string
	StripeToken       string
}

func UpdatePayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var req UpdatePaymentRequest
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

	customer.List(&stripe.CustomerListParams{})

	allCustomers := customer.List(&stripe.CustomerListParams{})

	var thisCustomer *stripe.Customer
	for allCustomers.Next() {
		c := allCustomers.Customer()
		userid := c.Meta[string(GithubUserIDMetadata)]
		if userid == githubUid {
			thisCustomer = c
			break
		}
	}

	if thisCustomer == nil {
		// Github login, no associated stripe, that's an error
		//userErr("Please create your stripe association first. Alternately, consistency error")
		return
	}

	customerParams := &stripe.CustomerParams{}
	err = customerParams.SetSource(req.StripeToken)
	// TODO, check if error is 400 or 500
	if err != nil {
		logrus.Infof("Invalid stripe token: %v", err)
		userErr("Invalid stripe token")
		return
	}

	_, err = customer.Update(thisCustomer.ID, customerParams)
	if err != nil {
		logrus.Warnf("Error updating customer: %v", err)
		serverErr("Error with stripe api for updating customer")
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"ok": 1}`))
}
