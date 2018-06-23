package api

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

type AddSubReq struct {
	PlanName          string
	PlanNum           int64
	GithubAccessToken string
	IdempotencyToken  string
}

func AddSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var req AddSubReq
	err := json.NewDecoder(r.Body).Decode(&req)
	r.Body.Close()
	if err != nil {
		userErr(w, "Invalid request")
		return
	}

	thisCustomer, err := getStripeUser(req.GithubAccessToken)
	if err != nil {
		userErr(w, err.Error())
		return
	}

	// find the existing sub for this plan if we can
	subs := sub.List(&stripe.SubscriptionListParams{Customer: thisCustomer.ID})
	if subs.Err() != nil {
		serverErr(w, "Error getting stripe subscriptions")
		return
	}

	var existing *stripe.Subscription
	for subs.Next() {
		sub := subs.Subscription()
		if sub.Status == stripe.SubscriptionStatusCanceled {
			continue
		}
		if sub.Plan.ID == req.PlanName {
			existing = sub
			break
		}
	}

	var params *stripe.SubscriptionParams
	var newSub *stripe.Subscription
	if existing == nil {
		// Create a new subscription
		params = &stripe.SubscriptionParams{
			Plan:     &req.PlanName,
			Quantity: &req.PlanNum,
			Customer: &thisCustomer.ID,
		}
		if req.IdempotencyToken != "" {
			params.IdempotencyKey = &req.IdempotencyToken
		}
		newSub, err = sub.New(params)
	} else {
		// One already exists, update the quantity
		newQuantity := existing.Quantity + req.PlanNum
		params = &stripe.SubscriptionParams{
			Quantity: &newQuantity,
		}
		if req.IdempotencyToken != "" {
			params.IdempotencyKey = &req.IdempotencyToken
		}

		newSub, err = sub.Update(existing.ID, params)
	}
	if err != nil {
		// Scary!
		logrus.Errorf("Error creating a subscription: %v, %v", params, err)
		serverErr(w, "Something went wrong making a subscription: "+err.Error())
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"Ok":1,"SubscriptionID":"` + newSub.ID + `"}`))
}
