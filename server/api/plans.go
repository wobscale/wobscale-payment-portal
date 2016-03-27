package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

type Plan string

const (
	Plan1U = "colo-sea1-1U"
	Plan2U = "colo-sea1-2U"
)

func PlanPrice(plan string) (uint64, error) {
	_, plans, err := getStripePlans()
	if err != nil {
		return 0, err
	}

	for _, p := range plans {
		if p.ID == plan {
			return p.Amount, nil
		}
	}

	return 0, errors.New("Invalid plan")
}

func ValidPlan(plan string) bool {
	_, err := PlanPrice(plan)
	return err == nil
}

type SubPlan struct {
	Name string
	Cost uint64
	Num  uint64
}

type PlanResp []PlanRespEl

type PlanRespEl struct {
	Name string
	ID   string
	Cost uint64
}

var stripePlanCacheMutex sync.Mutex
var stripePlanCache []*stripe.Plan
var stripePlanRespCache []byte

func getStripePlans() ([]byte, []*stripe.Plan, error) {
	stripePlanCacheMutex.Lock()
	defer stripePlanCacheMutex.Unlock()
	if stripePlanCache != nil {
		return stripePlanRespCache, stripePlanCache, nil
	}

	plans := plan.List(&stripe.PlanListParams{})
	if plans.Err() != nil {
		return nil, nil, errors.New("Error getting plans: " + plans.Err().Error())
	}

	planArr := []*stripe.Plan{}
	planResp := PlanResp{}
	for plans.Next() {
		plan := plans.Plan()
		planArr = append(planArr, plan)
		planResp = append(planResp, PlanRespEl{
			Name: plan.Name,
			ID:   plan.ID,
			Cost: plan.Amount,
		})
	}

	tmp, err := json.Marshal(planResp)
	if err != nil {
		return nil, nil, err
	}

	stripePlanCache = planArr
	stripePlanRespCache = tmp
	return stripePlanRespCache, stripePlanCache, nil
}

func GetPlans(w http.ResponseWriter, r *http.Request) {
	data, _, err := getStripePlans()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	logrus.Info(string(data))
	w.WriteHeader(200)
	w.Write(data)
}
