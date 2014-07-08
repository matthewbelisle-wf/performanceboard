package performanceboard

import (
	"net/http"
	"net/url"
)

func AbsURL(u *url.URL, request *http.Request) string {
	if !u.IsAbs() {
		if request.TLS == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
		u.Host = request.Host
	}
	return u.String()
}
