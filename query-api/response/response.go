/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    response.go
 * details: Deals with the response handlers for all DB handler
 *
 */
package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	opts "github.com/Juniper/collector/query-api/options"
)

func EncodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func Respond(w http.ResponseWriter, r *http.Request, status int,
	data interface{}) {
	if opts.Verbose {
		opts.Logger.Println("Respond() from Query API Server ", status)
	}
	w.WriteHeader(status)
	if data != nil {
		EncodeBody(w, r, data)
	}
}

func RespondErr(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	errMsg := fmt.Sprint(args...)
	opts.Logger.Println("RespondErr() from Query API Server ", status, errMsg)
	Respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": errMsg,
		},
	})
}

func RespondHTTPErr(w http.ResponseWriter, r *http.Request, status int) {
	opts.Logger.Println("RespondHTTPErr() from Query API Server ", status, http.StatusText(status))
	RespondErr(w, r, status, http.StatusText(status))
}
