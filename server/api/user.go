package api

import (
	"encoding/json"
	"net/http"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/card"
	"github.com/stripe/stripe-go/sub"
)

type GetUserReq struct {
	GithubAccessToken string
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	var req GetUserReq
	err := json.NewDecoder(r.Body).Decode(&req)
	r.Body.Close()
	if err != nil {
		userErr(w, "Malformed request")
		return
	}

	thisCustomer, authedUser, err := getStripeAndGithubUser(req.GithubAccessToken)

	if err == NoSuchCustomerErr {
		writeHappyResp(w, GetUserNewUserResp{
			GithubUsername: *authedUser.Login,
			NewUser:        true,
		})
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

	custCard, err := card.Get(thisCustomer.DefaultSource.ID, &stripe.CardParams{Customer: thisCustomer.ID})
	if err != nil {
		serverErr(w, "Difficulty getting payment source")
		return
	}

	cardStr := custCard.Display()
	resp := GetUserResp{
		GithubUsername:   *authedUser.Login,
		StripeCustomerID: thisCustomer.ID,
		PaymentSource:    cardStr,
		Plans:            subPlans,
	}

	writeHappyResp(w, resp)
}

type GetUserResp struct {
	GithubUsername   string
	StripeCustomerID string
	Plans            []SubPlan
	PaymentSource    string
}

type GetUserNewUserResp struct {
	GithubUsername string
	NewUser        bool
}
