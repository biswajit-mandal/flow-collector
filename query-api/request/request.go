/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    request.go
 * details: Deals with the validity of the request or any other processing of the request before
 *          passing to the actual handler
 *
 */
package request

import (
	"encoding/json"
	"net/http"

	res "github.com/Juniper/collector/query-api/response"
)

func DecodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func DecodeBodyNumber(r *http.Request, v interface{}) error {
	d := json.NewDecoder(r.Body)
	d.UseNumber()
	return d.Decode(v)
}

func IsValidAuthKey(key string) bool {
	/* We need to validate the request */
	return true
}

func IsValidRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsValidAuthKey(r.URL.Query().Get("X-Auth-Token")) {
			res.RespondErr(w, r, http.StatusUnauthorized, "Invalid Auth Key")
			return
		}
		fn(w, r)
	}
}
