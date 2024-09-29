package middleware

import (
	"net/http"
	"strings"
	"time"
)

func RateLimit(next http.Handler) http.Handler {
	ips := map[string]time.Time{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			addr := r.Header.Get("X-Forwarded-For")

			xffSplits := strings.Split(addr, ",")
			xffStr := ""
			if len(xffSplits) > 0 {
				xffStr = xffSplits[len(xffSplits)-1]
			}

			if xffStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			ts, exists := ips[xffStr]
			ips[xffStr] = time.Now()

			if exists && time.Since(ts) < time.Minute {
				time.Sleep(time.Second * 30)
				w.WriteHeader(420)
				w.Write([]byte("Enhance your calm"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func IPBlock(ips []string, next http.Handler) http.Handler {
	ipList := map[string]struct{}{}
	for _, ip := range ips {
		ipList[ip] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr := r.Header.Get("X-Forwarded-For")

		xffSplits := strings.Split(addr, ",")
		xffStr := ""
		if len(xffSplits) > 0 {
			xffStr = xffSplits[len(xffSplits)-1]
		}

		_, exists := ipList[xffStr]

		if exists {
			time.Sleep(5 * time.Minute)
			w.WriteHeader(http.StatusTeapot)
			return
		}

		next.ServeHTTP(w, r)
	})
}
