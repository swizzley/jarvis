package util

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/blendlabs/go-exception"
)

// Write an object to a response as json
func WriteJson(w http.ResponseWriter, statusCode int, response interface{}) (int, error) {
	bytes, err := json.Marshal(response)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		count, write_error := w.Write(bytes)
		if count == 0 {
			return count, exception.New("WriteJson : Didnt write any bytes.")
		}
		return count, write_error
	} else {
		return 0, err
	}
}

// Getting the correct IP is not necessarily straightforward but this is the best attempt
// X-FORWARDED-FOR is checked. If multiple IPs are included the first one is returned
// X-REAL-IP is checked. If multiple IPs are included the first one is returned
// Finally r.RemoteAddr is used
// Only benevolent services will allow access to the real IP
func GetIP(r *http.Request) string {
	tryHeader := func(key string) (string, bool) {
		if headerVal := r.Header.Get(key); len(headerVal) > 0 {
			if !strings.ContainsRune(headerVal, ',') {
				return headerVal, true
			}
			return strings.SplitN(headerVal, ",", 2)[0], true
		}
		return "", false
	}

	for _, header := range []string{"X-FORWARDED-FOR", "X-REAL-IP"} {
		if headerVal, ok := tryHeader(header); ok {
			return headerVal
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func GetParamByName(r *http.Request, name string) string {
	//check querystring
	query_value := r.URL.Query().Get(name)
	if !IsEmpty(query_value) {
		return query_value
	}

	//check headers
	header_value := r.Header.Get(name)
	if !IsEmpty(header_value) {
		return header_value
	}

	//check cookies
	cookie, cookie_err := r.Cookie(name)
	if cookie_err == nil && !IsEmpty(cookie.Value) {
		return cookie.Value
	}

	form_value := r.Form.Get(name)
	if !IsEmpty(form_value) {
		return form_value
	}

	post_form_value := r.PostFormValue(name)
	if !IsEmpty(post_form_value) {
		return post_form_value
	}

	return EMPTY
}
