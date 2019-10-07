package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func SecondsSince(start time.Time) float64 {
	return float64(time.Now().Sub(start).Seconds())
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

func ParseSinceParameter(since string) (*time.Time, error) {
	// storing since we mutate `since`
	_since := since

	matcher, err := regexp.Compile("([0-9]+d)?([0-9]+h)?([0-9]+m)?([0-9]+s)?")
	if err != nil {
		log.Printf("ERROR: cannot compile regexp in `ParseSinceParameter`\n")
		return nil, err
	}
	match := matcher.MatchString(since)
	if !match {
		log.Printf("ERROR: cannot find pattern in `ParseSinceParameter`; passed %s\n", _since)
		return nil, fmt.Errorf("Invalid parameter provided: %s", _since)
	}

	matches := matcher.FindAllString(since, -1)

	if len(matches) > 1 {
		log.Printf("ERROR: cannot find pattern in `ParseSinceParameter`; passed %s\n", _since)
		return nil, fmt.Errorf("Invalid parameter provided: %s", _since)
	}

	// parse days, hours, minutes, seconds
	var days, hours, mins, secs int64
	sinceSplit := strings.Split(since, "d")
	if len(sinceSplit) > 1 {
		days, err = strconv.ParseInt(sinceSplit[0], 10, 32)
		if err != nil {
			log.Printf("ERROR: cannot parse `d` days in `ParseSinceParameter`; passed %s\n", _since)
			return nil, fmt.Errorf("Unable to parse `d` days; invalid parameter provided: %s", _since)
		}
		since = sinceSplit[1]
	}
	sinceSplit = strings.Split(since, "h")
	if len(sinceSplit) > 1 {
		hours, err = strconv.ParseInt(sinceSplit[0], 10, 32)
		if err != nil {
			log.Printf("ERROR: cannot parse `h` hours in `ParseSinceParameter`; passed %s\n", _since)
			return nil, fmt.Errorf("Unable to parse `h` hours, invalid parameter provided: %s", _since)
		}
		since = sinceSplit[1]
	}
	sinceSplit = strings.Split(since, "m")
	if len(sinceSplit) > 1 {
		mins, err = strconv.ParseInt(sinceSplit[0], 10, 32)
		if err != nil {
			log.Printf("ERROR: cannot parse `m` minutes in `ParseSinceParameter`; passed %s\n", _since)
			return nil, fmt.Errorf("Unable to parse `m` minutes, invalid parameter provided: %s", _since)
		}
		since = sinceSplit[1]
	}
	sinceSplit = strings.Split(since, "s")
	if len(sinceSplit) > 1 {
		secs, err = strconv.ParseInt(sinceSplit[0], 10, 32)
		if err != nil {
			log.Printf("ERROR: cannot parse `s` seconds in `ParseSinceParameter`; passed %s\n", _since)
			return nil, fmt.Errorf("Unable to parse `s` seconds, invalid parameter provided: %s", _since)
		}
		since = sinceSplit[1]
	}

	startDate := time.Now().UTC().
		Add(time.Duration(-secs) * time.Second).
		Add(time.Duration(-mins) * time.Minute).
		Add(time.Duration(-hours+(-days*24)) * time.Hour)

	return &startDate, nil

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

func RespondWithJSON(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func StructToFile(in interface{}) {
	data, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("output.json", data, 0644)
	if err != nil {
		panic(err)
	}
}

func StructFromFile(out interface{}) {
	data, err := ioutil.ReadFile("output.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, out)
	if err != nil {
		panic(err)
	}
}
