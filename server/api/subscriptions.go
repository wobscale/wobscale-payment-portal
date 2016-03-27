package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
	"golang.org/x/oauth2"
)

type AddSubReq struct {
	PlanName          string
	PlanNum           uint64
	GithubAccessToken string
	IdempotencyToken  string
}

func AddSubscription(w http.ResponseWriter, r *http.Request) {
	serverErr := func(err string) {
		w.WriteHeader(500)
		out, _ := json.Marshal(apiErr{err})
		w.Write(out)
	}

	var req AddSubReq
	err := json.NewDecoder(r.Body).Decode(&req)
	r.Body.Close()
	if err != nil {
		// TODO
		return
	}
	githubauth := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: req.GithubAccessToken})
	githurb := github.NewClient(oauth2.NewClient(oauth2.NoContext, githubauth))
	authedUser, _, err := githurb.Users.Get("")
	if err != nil {
		logrus.Errorf("Unable to get currently logged in github user: %v", err)
		w.WriteHeader(400)
		w.Write([]byte(`{"GithubAuthError": true}`))
		return
	}

	if authedUser.Login == nil {
		logrus.Errorf("Nil github username: %v", authedUser)
		//
		return
	}
	if authedUser.ID == nil {
		logrus.Errorf("Nil userid: %v", authedUser)
		// TODO
		return
	}

	// Search stripe for this sucker
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

	// find the existing sub for this plan if we can
	subs := sub.List(&stripe.SubListParams{Customer: thisCustomer.ID})
	if subs.Err() != nil {
		serverErr("Error getting stripe subscriptions")
		return
	}

	var existing *stripe.Sub
	for subs.Next() {
		sub := subs.Sub()
		if sub.Plan.ID == req.PlanName {
			existing = sub
			break
		}
	}

	var params *stripe.SubParams
	var newSub *stripe.Sub
	if existing == nil {
		// Create a new subscription
		params = &stripe.SubParams{
			Plan:     req.PlanName,
			Quantity: req.PlanNum,
			Customer: thisCustomer.ID,
		}
		params.IdempotencyKey = req.IdempotencyToken
		newSub, err = sub.New(params)
	} else {
		// One already exists, update the quantity
		params = &stripe.SubParams{
			Plan:     req.PlanName,
			Quantity: existing.Quantity + req.PlanNum,
			Customer: thisCustomer.ID,
		}

		newSub, err = sub.Update(existing.ID, params)
	}
	if err != nil {
		// Scary!
		logrus.Errorf("Error creating a subscription: %v, %v", params, err)
		serverErr("Something went wrong making a subscription: " + err.Error())
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"Ok":1,"SubscriptionID":"` + newSub.ID + `"}`))
}
