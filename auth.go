// Copyright 2014 GWoo. All rights reserved.
// The BSD License http://opensource.org/licenses/bsd-license.php.
package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Auth struct {
	Realm string
	Creds string
}

func (a Auth) NotAuthorized(url string, w http.ResponseWriter) {
	log.Printf("Unauthorized access to %s", url)
	w.Header().Add("WWW-Authenticate", fmt.Sprintf("basic realm=\"%s\"", a.Realm))
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "Not Authorized.")
}

func AuthHandler(auth Auth, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		header, ok := r.Header["Authorization"]
		if !ok {
			auth.NotAuthorized(url.String(), w)
			return
		}
		encoded := strings.Split(header[0], " ")
		if len(encoded) != 2 || encoded[0] != "Basic" {
			log.Printf("Strange Authorization %q", header)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded[1])
		if err != nil {
			log.Printf("Cannot decode %q: %s", header, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if auth.Creds == string(decoded) {
			handler.ServeHTTP(w, r)
			return
		}
		auth.NotAuthorized(url.String(), w)
		return
	})
}
