// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released
// under the terms of the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for
// details.  Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package napping

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// A Params is a map containing URL parameters.
type Params map[string]string

// AsUrlValues converts Params to url.Values
func (p Params) AsUrlValues() url.Values {
	result := url.Values{}
	for key, value := range p {
		result.Set(key, value)
	}
	return result
}

// A Request describes an HTTP request to be executed, data structures into
// which the result will be unmarshaled, and the server's response. By using
// a  single object for both the request and the response we allow easy access
// to Result and Error objects without needing type assertions.
type Request struct {
	Url     string      // Raw URL string
	Method  string      // HTTP method to use
	Params  *url.Values // URL query parameters
	Payload interface{} // Data to JSON-encode and POST

	// Not capture response body and unmarshaled
	NotProcessBody bool

	// Optional
	Userinfo *url.Userinfo
	Header   *http.Header

	// Custom Transport if needed.
	Transport *http.Transport

	// The following fields are populated by Send().
	timestamp time.Time      // Time when HTTP request was sent
	status    int            // HTTP status for executed request
	response  *http.Response // Response object from http package
	body      []byte         // Body of server's response (JSON or otherwise)
}

// A Response is a Request object that has been executed.
type Response Request

// Timestamp returns the time when HTTP request was sent.
func (r *Response) Timestamp() time.Time {
	return r.timestamp
}

// Timestamp returns the time when HTTP request was sent.
func (r *Response) RawByte() []byte {
	return r.body
}

// RawText returns the body of the server's response as raw text.
func (r *Response) RawText() string {
	return strings.TrimSpace(string(r.body))
}

// Status returns the HTTP status for the executed request, or 0 if request has
// not yet been sent.
func (r *Response) Status() int {
	return r.status
}

func (r *Response) StatusOk() bool {
	return r.status <= 300
}

func (r *Response) IsJsonMime() bool {
	if r.response == nil {
		return false
	}
	contentType := r.response.Header.Get("Content-Type")
	isJson := strings.HasSuffix(strings.Split(contentType, ";")[0], "json")
	return isJson
}

// HttpResponse returns the underlying Response object from http package.
func (r *Response) HttpResponse() *http.Response {
	return r.response
}

// Unmarshal parses the JSON-encoded data in the server's response, and stores
// the result in the value pointed to by v.
func (r *Response) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.body, v)
}
