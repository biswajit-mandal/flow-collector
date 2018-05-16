/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    dbhandler.go
 * details: Initializes all the database handlers
 *
 */
package dbhandler

import (
	"net/http"

	opts "github.com/Juniper/collector/query-api/options"
)

type Handler struct {
	dbH dbHandler
}

type dbHandler interface {
	setup(*http.ServeMux) error
}

func NewDBHandler(handlerName string) *Handler {
	var dbHandlerRegistered = map[string]dbHandler{
		opts.UseDatabaseMongo: new(MongoDBHandler),
	}
	return &Handler{
		dbH: dbHandlerRegistered[handlerName],
	}
}

func (h Handler) Run(mux *http.ServeMux) error {
	var (
		err error
	)
	err = h.dbH.setup(mux)
	return err
}
