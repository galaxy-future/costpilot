package tools

import (
	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"go.uber.org/ratelimit"
)

var Limiters map[string]ratelimit.Limiter

func NewLimiters() {
	Limiters = make(map[string]ratelimit.Limiter)
	Limiters[cloud.TencentCloud+"-"+"DescribeBillDetail"] = ratelimit.New(3, ratelimit.WithoutSlack)
	//Limiters[cloud.AWSCloud]
}
