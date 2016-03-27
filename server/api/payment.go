package api

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
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
	if err != nil {
		userErr(w, "Malformed request")
		return
	}

	if req.GithubAccessToken == "" {
		userErr(w, "Github auth is required")
		return
	}

	if req.StripeToken == "" {
		userErr(w, "Stripe token is required")
		return
	}

	thisCustomer, err := getStripeUser(req.GithubAccessToken)
	if err != nil {
		userErr(w, err.Error())
		return
	}

	customerParams := &stripe.CustomerParams{}
	err = customerParams.SetSource(req.StripeToken)
	// TODO, check if error is 400 or 500
	if err != nil {
		logrus.Infof("Invalid stripe token: %v", err)
		userErr(w, "Invalid stripe token")
		return
	}

	_, err = customer.Update(thisCustomer.ID, customerParams)
	if err != nil {
		logrus.Warnf("Error updating customer: %v", err)
		serverErr(w, "Error with stripe api for updating customer")
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"ok": 1}`))
}
