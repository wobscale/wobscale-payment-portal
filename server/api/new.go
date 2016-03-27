package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

type CustomerMetadata string

const (
	GithubUsernameMetadata CustomerMetadata = "GithubUsername"
	GithubUserIDMetadata   CustomerMetadata = "GithubUserID"
)

type SubscriptionRequest struct {
	Nickname          string
	GithubAccessToken string
	Email             string
	StripeToken       string
}

func NewSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var req SubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	r.Body.Close()

	if err != nil {
		userErr(w, "Malformed request")
		return
	}

	if req.GithubAccessToken == "" {
		userErr(w, "Github auth is required")
		return
	}

	if req.Email == "" {
		userErr(w, "Email is required")
		return
	}

	if req.StripeToken == "" {
		userErr(w, "Stripe setup (CC etc) is required")
		return
	}

	authedUser, err := githubUser(req.GithubAccessToken)
	if err != nil {
		// TODO clasify
		userErr(w, err.Error())
		return
	}
	// Validate githurb stuff correctly now
	githubUid := strconv.Itoa(*authedUser.ID)

	// valid, let's create this customer

	customerParams := &stripe.CustomerParams{
		Email: req.Email,
		Desc:  req.Nickname,
	}

	customerParams.Meta = map[string]string{
		string(GithubUserIDMetadata):   githubUid,
		string(GithubUsernameMetadata): *authedUser.Login,
	}
	err = customerParams.SetSource(req.StripeToken)
	// TODO, check if error is 400 or 500
	if err != nil {
		logrus.Infof("Invalid stripe token: %v", err)
		userErr(w, "Invalid stripe token")
		return
	}

	createdCustomer, err := customer.New(customerParams)
	if err != nil {
		logrus.Warnf("Error creating customer: %v", err)
		serverErr(w, "Error with stripe api for creating customer")
		return
	}
	logrus.Infof("Customer created: %v, %v", req.Email, createdCustomer.ID)

	// TODO, email admin@ or something?
	// TODO, optimize by returning customer-id here, adding a getter that takes it

	w.WriteHeader(200)
	w.Write([]byte(`{"ok": 1}`))
}
