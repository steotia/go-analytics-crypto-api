package validators

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

func CheckParam(r *http.Request, param string) (string, error) {
	params, ok := r.URL.Query()[param]
	if !ok || len(params[0]) < 1 {
		message := fmt.Sprintf("Can't process: Url Param '%s' is missing", param)
		return "", errors.New(message)
	}
	return params[0], nil
}

func CheckTimeFormat(k string, v string) (time.Time, error) {
	vTime, err := time.Parse(timeFormat, v)
	if err != nil {
		message := fmt.Sprintf("Can't process: '%s' not in format: '2006-01-02T15:04:05'", k)
		return time.Time{}, errors.New(message)
	}
	return vTime, nil
}
