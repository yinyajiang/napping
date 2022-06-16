// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released
// under the terms of the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for
// details.  Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package napping

/*
This module provides a Session object to manage and persist settings across
requests (cookies, auth, proxies).
*/

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Session defines the napping session structure
type Session struct {
	Client *http.Client

	// Optional
	Userinfo *url.Userinfo

	// Optional defaults - can be overridden in a Request
	Header *http.Header
	Params *url.Values
}

// Send constructs and sends an HTTP request.
func (s *Session) Send(r *Request) (response *Response, err error) {
	r.Method = strings.ToUpper(r.Method)

	// Create a URL object from the raw url string.  This will allow us to compose
	// query parameters programmatically and be guaranteed of a well-formed URL.
	u, err := url.Parse(r.Url)
	if err != nil {
		s.log("URL", r.Url)
		s.log(err)
		return
	}

	// Default query parameters
	p := url.Values{}
	if s.Params != nil {
		for k, v := range *s.Params {
			p[k] = v
		}
	}

	// Parameters that were present in URL
	if u.Query() != nil {
		for k, v := range u.Query() {
			p[k] = v
		}
	}

	// User-supplied params override default
	if r.Params != nil {
		for k, v := range *r.Params {
			p[k] = v
		}
	}

	// Encode parameters
	u.RawQuery = p.Encode()

	// Attach params to response
	r.Params = &p

	// Create a Request object; if populated, Data field is JSON encoded as request body
	header := http.Header{}
	if s.Header != nil {
		for k := range *s.Header {
			v := s.Header.Get(k)
			header.Set(k, v)
		}
	}

	var paylodReader io.Reader
	if r.Payload != nil {
		if _, ok := r.Payload.(io.Reader); ok {
			r.Payload = r.Payload.(io.Reader)
		} else {
			var bydata []byte
			kind := reflect.TypeOf(r.Payload).Kind()
			switch kind {
			case reflect.Map:
				fallthrough
			case reflect.Struct:
				bydata, err = json.Marshal(r.Payload)
			case reflect.String:
				r.Payload = []byte(r.Payload.(string))
				fallthrough
			case reflect.Slice:
				var ok bool
				bydata, ok = r.Payload.([]byte)
				if !ok {
					jsArr, ok := r.Payload.([]interface{})
					if ok {
						bydata, _ = json.Marshal(jsArr)
					}
				}
			}
			if err != nil {
				return
			}
			if len(bydata) != 0 {
				paylodReader = bytes.NewBuffer(bydata)
				if ("{" == string(bydata[0]) && "}" == string(bydata[len(bydata)-1])) ||
					("[" == string(bydata[0]) && "]" == string(bydata[len(bydata)-1])) {
					header.Set("Content-Type", "application/json")
				}
			}
		}
	}

	req, err := http.NewRequest(r.Method, u.String(), paylodReader)
	if err != nil {
		s.log(err)
		return
	}

	// Merge Session and Request options
	var userinfo *url.Userinfo
	if u.User != nil {
		userinfo = u.User
	}
	if s.Userinfo != nil {
		userinfo = s.Userinfo
	}
	// Prefer Request's user credentials
	if r.Userinfo != nil {
		userinfo = r.Userinfo
	}
	if r.Header != nil {
		for k, v := range *r.Header {
			header.Set(k, v[0]) // Is there always guarnateed to be at least one value for a header?
		}
	}
	if header.Get("Accept") == "" {
		header.Add("Accept", "*/*") // Default, can be overridden with Opts
	}
	req.Header = header

	// Set HTTP Basic authentication if userinfo is supplied
	if userinfo != nil {
		pwd, _ := userinfo.Password()
		req.SetBasicAuth(userinfo.Username(), pwd)
		if u.Scheme != "https" {
			s.log("WARNING: Using HTTP Basic Auth in cleartext is insecure.")
		}
	}

	r.timestamp = time.Now()
	var client *http.Client
	if s.Client != nil {
		client = s.Client
	} else {
		client = &http.Client{}
		if r.Transport != nil {
			client.Transport = r.Transport
		}

		s.Client = client
	}
	resp, err := client.Do(req)
	if err != nil {
		s.log(err)
		return
	}
	r.status = resp.StatusCode
	r.response = resp

	if !r.NotProcessBody {
		defer resp.Body.Close()

		// Unmarshal
		r.body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			s.log(err)
			return
		}
		if string(r.body) != "" {
			if resp.StatusCode <= 200 && r.Result != nil {
				json.Unmarshal(r.body, r.Result)
			}
			if resp.StatusCode > 200 && r.Error != nil {
				json.Unmarshal(r.body, r.Error) // Should we ignore unmarshal error?
			}
		}
	}

	rsp := Response(*r)
	response = &rsp
	return
}

// Get sends a GET request.
func (s *Session) Get(url string, p *url.Values, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method: "GET",
		Url:    url,
		Params: p,
		Result: result,
		Error:  errMsg,
	}
	return s.Send(&r)
}

// Options sends an OPTIONS request.
func (s *Session) Options(url string, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method: "OPTIONS",
		Url:    url,
		Result: result,
		Error:  errMsg,
	}
	return s.Send(&r)
}

// Head sends a HEAD request.
func (s *Session) Head(url string, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method: "HEAD",
		Url:    url,
		Result: result,
		Error:  errMsg,
	}
	return s.Send(&r)
}

// Post sends a POST request.
func (s *Session) Post(url string, payload, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method:  "POST",
		Url:     url,
		Payload: payload,
		Result:  result,
		Error:   errMsg,
	}
	return s.Send(&r)
}

// Put sends a PUT request.
func (s *Session) Put(url string, payload, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method:  "PUT",
		Url:     url,
		Payload: payload,
		Result:  result,
		Error:   errMsg,
	}
	return s.Send(&r)
}

// Patch sends a PATCH request.
func (s *Session) Patch(url string, payload, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method:  "PATCH",
		Url:     url,
		Payload: payload,
		Result:  result,
		Error:   errMsg,
	}
	return s.Send(&r)
}

// Delete sends a DELETE request.
func (s *Session) Delete(url string, p *url.Values, result, errMsg interface{}) (*Response, error) {
	r := Request{
		Method: "DELETE",
		Url:    url,
		Params: p,
		Result: result,
		Error:  errMsg,
	}
	return s.Send(&r)
}

// Debug method for logging
// Centralizing logging in one method
// avoids spreading conditionals everywhere
func (s *Session) log(args ...interface{}) {
	log.Println(args...)
}
