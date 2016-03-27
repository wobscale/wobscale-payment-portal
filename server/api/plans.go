package api

import "errors"

type Plan string

const (
	Plan1U = "1U"
	Plan2U = "2U"
)

func ValidPlan(plan string) bool {
	return plan == Plan1U || plan == Plan2U
}

func PlanPrice(plan string) (int64, error) {
	switch plan {
	case Plan1U:
		return 80, nil
	case Plan2U:
		return 140, nil
	default:
		return 0, errors.New("Invalid plan")
	}
}
