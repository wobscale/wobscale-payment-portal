package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Plan string

const (
	Plan1U = "colo-sea1-1U"
	Plan2U = "colo-sea1-2U"
)

func ValidPlan(plan string) bool {
	return plan == Plan1U || plan == Plan2U
}

func PlanPrice(plan string) (uint64, error) {
	switch plan {
	case Plan1U:
		return 80, nil
	case Plan2U:
		return 140, nil
	default:
		return 0, errors.New("Invalid plan")
	}
}

type SubPlan struct {
	Name string
	Cost uint64
	Num  uint64
}

type PlanResp []PlanRespEl

type PlanRespEl struct {
	Name string
	Cost uint64
}

var planResp PlanResp
var planRespCache []byte

func init() {
	var err error
	planResp = PlanResp{}
	for _, el := range []string{Plan1U, Plan2U} {
		price, err := PlanPrice(el)
		if err != nil {
			panic(err)
		}
		planResp = append(planResp, PlanRespEl{Name: el, Cost: price})
	}
	planRespCache, err = json.Marshal(planResp)
	if err != nil {
		panic(err)
	}
}

func GetPlans(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write(planRespCache)
}
