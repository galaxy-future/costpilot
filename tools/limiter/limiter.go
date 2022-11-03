package limiter

import (
	"sync"

	"go.uber.org/ratelimit"
)

type LimiterSet map[string]ratelimit.Limiter

var (
	Limiters LimiterSet
	lock     sync.Mutex
)

func init() {
	Limiters = make(map[string]ratelimit.Limiter)
}
func (l *LimiterSet) GetLimiter(key string, rate int) ratelimit.Limiter {
	lock.Lock()
	defer lock.Unlock()
	limiter, ok := Limiters[key]
	if ok {
		return limiter
	}
	Limiters[key] = ratelimit.New(rate, ratelimit.WithoutSlack)
	return Limiters[key]
}
