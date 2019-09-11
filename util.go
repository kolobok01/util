package util

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"
)

func SecondsSince(start time.Time) float64 {
	return float64(time.Now().Sub(start).Seconds())
}

func JsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"value": map[string]string{
				"message": msg,
			},
			"status": 13,
		})
}

const UnknownUser = "unknown"

func RequestInfo(r *http.Request) (string, string) {
	user := ""
	if u, _, ok := r.BasicAuth(); ok {
		user = u
	} else {
		user = UnknownUser
	}
	remote := r.Header.Get("X-Forwarded-For")
	if remote != "" {
		return user, remote
	}
	remote, _, _ = net.SplitHostPort(r.RemoteAddr)
	return user, remote
}

func HostPort(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		panic(err)
	}
	return u.Host
}
