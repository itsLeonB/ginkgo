package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo/pkg/response"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func newRateLimiter(r rate.Limit, b int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *rateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// NewRateLimitMiddleware creates a rate limiter middleware for Gin.
// It limits requests based on the client's IP address using a token bucket algorithm.
// limit: The number of requests per second derived from time.Duration (e.g., 1 request per second).
// burst: The maximum number of requests allowed to exceed the limit.
func (mp *MiddlewareProvider) NewRateLimitMiddleware(limit rate.Limit, burst int) gin.HandlerFunc {
	rl := newRateLimiter(limit, burst)

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			mp.logger.Warnf("rate limit exceeded for IP: %s", ip)
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, response.NewErrorResponse(errorObject{
				Code:   http.StatusText(http.StatusTooManyRequests),
				Detail: "rate limit exceeded",
			}))
			return
		}

		ctx.Next()
	}
}
