package middleware

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Period  time.Duration
	MaxRate int64
	Store   *redis.Client
}

var re = regexp.MustCompile(`\s?,\s?`)

func (rl *RateLimiter) writeRateLimitHeaders(
	w http.ResponseWriter,
	used int64,
	expireTime time.Duration,
) {
	limit := rl.MaxRate
	remaining := int64(math.Max(float64(limit-used), 0))
	reset := int64(math.Ceil(expireTime.Seconds()))

	w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
	w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtain the clientIP from the XFF header
		clientIP := re.Split(r.Header.Get("X-Forwarded-For"), -1)[0]

		// If the xff header is empty, obtain the IP from the remoteAddr
		if clientIP == "" {
			parts := strings.Split(r.RemoteAddr, ":")
			clientIP = strings.Join(parts[0:len(parts)-1], ":")
		}

		// Get the current time to use for the event
		now := time.Now()

		// Add the current event to the store
		rl.Store.ZAdd(r.Context(), clientIP, redis.Z{
			Member: now.UnixMicro(),
			Score:  float64(now.UnixMicro()),
		})

		// Calculate the cutoff
		cutoff := now.Add(rl.Period * -1).UnixMicro()

		// Remove all events that are before the cutoff
		rl.Store.ZRemRangeByScore(r.Context(), clientIP, "-inf", strconv.FormatInt(cutoff, 10))

		// Pull the remaining events from the sorted set
		events, _ := rl.Store.ZRange(r.Context(), clientIP, 0, -1).Result()

		// Get the earliest event time
		earliestMicro, _ := strconv.ParseInt(events[0], 10, 64)
		earliest := time.UnixMicro(earliestMicro)

		// Calculate how long until it resets and how many events have occurred
		resets := rl.Period - time.Since(earliest)
		eventCount := int64(len(events))

		// write the rate limit headers
		rl.writeRateLimitHeaders(w, eventCount, resets)

		// Check if client has exceeded the max rate
		if eventCount > rl.MaxRate {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		// Call the next handler if rate is not exceeded
		next.ServeHTTP(w, r)
	})
}
