package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"golang.org/x/oauth2"
)

type apiErr struct {
	Error string
}

func userErr(w http.ResponseWriter, err string) {
	w.WriteHeader(400)
	out, _ := json.Marshal(apiErr{err})
	w.Write(out)
}

func serverErr(w http.ResponseWriter, err string) {
	w.WriteHeader(500)
	out, _ := json.Marshal(apiErr{err})
	w.Write(out)
}

func githubUser(token string) (*github.User, error) {
	githubauth := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	githurb := github.NewClient(oauth2.NewClient(oauth2.NoContext, githubauth))
	authedUser, _, err := githurb.Users.Get("")
	if err != nil {
		logrus.Info("Error with user's github auth key: ", err)
		return nil, err
	}
	if authedUser.Login == nil {
		logrus.Errorf("Nil github user name: %v", authedUser)
		return nil, errors.New("Nil username")
	}
	if authedUser.ID == nil {
		logrus.Errorf("Nil userid: %v", authedUser)
		return nil, errors.New("Nil userid")
	}
	return authedUser, nil
}

func getStripeUser(githubToken string) (*stripe.Customer, error) {
	githubUser, err := githubUser(githubToken)
	if err != nil {
		return nil, err
	}
	githubUid := strconv.Itoa(*githubUser.ID)
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
		return nil, errors.New("No such customer")
	}
	return thisCustomer, nil
}
