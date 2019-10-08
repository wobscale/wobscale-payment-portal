package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/sqs/mux"
	"github.com/stripe/stripe-go"
	"github.com/wobscale/wobscale-payment-portal/server/api"
)

func main() {
	stripeAPIKey := os.Getenv("STRIPE_API_KEY")
	if stripeAPIKey == "" {
		panic("STRIPE_API_KEY environment variable required")
	}
	stripe.Key = stripeAPIKey

	githubSecretKey := os.Getenv("GITHUB_SECRET_KEY")
	if githubSecretKey == "" {
		panic("GITHUB_SECRET_KEY environment variable required")
	}
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	if githubClientID == "" {
		panic("GITHUB_CLIENT_ID environment variable required")
	}
	isDev := false
	if os.Getenv("ENV_ENVIRONMENT") == "dev" {
		isDev = true
	}

	origin := os.Getenv("CORS_ALLOW_ORIGIN")
	if origin == "" {
		origin = "https://pay.wobscale.website"
		logrus.Warnf("Origin not set; defaulting to " + origin)
	}
	logLevel := logrus.WarnLevel
	if isDev {
		logLevel = logrus.DebugLevel
	}
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		var err error
		logLevel, err = logrus.ParseLevel(lvl)
		if err != nil {
			panic("LOG_LEVEL invalid")
		}
	}
	logrus.SetLevel(logLevel)

	router := mux.NewRouter().StrictSlash(true)

	withCors := func(f http.Handler) http.Handler {
		return api.WithCorsHeaders(origin, f.ServeHTTP)
	}

	router.HandleFunc("/ping", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("pong"))
		rw.WriteHeader(200)
	})
	router.HandleFunc("/new", api.NewSubscription)
	router.HandleFunc("/updatePayment", api.UpdatePayment)
	router.HandleFunc("/addSubscription", api.AddSubscription)
	router.HandleFunc("/user", api.GetUser)
	router.HandleFunc("/githubLogin", api.GithubLogin(githubSecretKey, githubClientID))
	router.HandleFunc("/plans", api.GetPlans)
	log.Fatal(http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, withCors(router))))
}
