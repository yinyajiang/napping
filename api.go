// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released
// under the terms of the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for
// details.  Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package napping

/*
This module implements the Napping API.
*/

import (
	"net/url"
)

// Send composes and sends and HTTP request.
func Send(r *Request) (*Response, error) {
	s := Session{}
	return s.Send(r)
}

// Get sends a GET request.
func Get(url string, p *url.Values) (*Response, error) {
	s := Session{}
	return s.Get(url, p)
}

// Options sends an OPTIONS request.
func Options(url string) (*Response, error) {
	s := Session{}
	return s.Options(url)
}

// Head sends a HEAD request.
func Head(url string) (*Response, error) {
	s := Session{}
	return s.Head(url)
}

// Post sends a POST request.
func Post(url string, payload interface{}) (*Response, error) {
	s := Session{}
	return s.Post(url, payload)
}

// Put sends a PUT request.
func Put(url string, payload interface{}) (*Response, error) {
	s := Session{}
	return s.Put(url, payload)
}

// Patch sends a PATCH request.
func Patch(url string, payload interface{}) (*Response, error) {
	s := Session{}
	return s.Patch(url, payload)
}

// Delete sends a DELETE request.
func Delete(url string, p *url.Values) (*Response, error) {
	s := Session{}
	return s.Delete(url, p)
}
