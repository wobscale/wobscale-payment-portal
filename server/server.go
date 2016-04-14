package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sqs/mux"
	"github.com/stripe/stripe-go"
	"github.com/wobscale/wobscale-payment-portal/server/api"
)

func main() {
	stripeApiKey := os.Getenv("STRIPE_API_KEY")
	if stripeApiKey == "" {
		panic("STRIPE_API_KEY environment variable required")
	}
	stripe.Key = stripeApiKey

	githubSecretKey := os.Getenv("GITHUB_SECRET_KEY")
	if githubSecretKey == "" {
		panic("GITHUB_SECRET_KEY environment variable required")
	}
	githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	if githubClientId == "" {
		panic("GITHUB_CLIENT_ID environment variable required")
	}

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/new", api.NewSubscription)
	router.HandleFunc("/updatePayment", api.UpdatePayment)
	router.HandleFunc("/addSubscription", api.AddSubscription)
	router.HandleFunc("/user", api.GetUser)
	router.HandleFunc("/githubLogin", api.GithubLogin(githubSecretKey, githubClientId))
	router.HandleFunc("/plans", api.GetPlans)
	log.Fatal(http.ListenAndServe(":8080", router))
}
