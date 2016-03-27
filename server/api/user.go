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

type GetUserReq struct {
	GithubAccessToken string
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	var req GetUserReq
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
			//		userErr("You already exist; please login to manage instead")
			thisCustomer = c
			break
		}
	}

	if thisCustomer == nil {
		// Github login, no associated stripe, that's still fine, they just have to sub-up
		resp := GetUserNewUserResp{NewUser: true, GithubUsername: *authedUser.Login}
		respData, err := json.Marshal(resp)
		if err != nil {
			logrus.Error("Broken marshal: %v, %v", err, resp)
			return
		}
		w.WriteHeader(200)
		w.Write(respData)
		return
	}

	subs := sub.List(&stripe.SubListParams{Customer: thisCustomer.ID})

	subPlans := []SubPlan{}
	for subs.Next() {
		sub := subs.Sub()
		plan := SubPlan{
			Name: sub.Plan.Name,
			Cost: sub.Plan.Amount,
			Num:  sub.Quantity,
		}
		subPlans = append(subPlans, plan)
	}

	resp := GetUserResp{
		GithubUsername:   *authedUser.Name,
		StripeCustomerID: thisCustomer.ID,
		PaymentSource:    thisCustomer.DefaultSource,
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		logrus.Error("Broken marshal: %v, %v", err, resp)
		return
	}

	w.WriteHeader(200)
	w.Write(respData)
}

type GetUserResp struct {
	GithubUsername   string
	StripeCustomerID string
	Plans            []SubPlan
	PaymentSource    *stripe.PaymentSource
}

type GetUserNewUserResp struct {
	GithubUsername string
	NewUser        bool
}
