/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    handler.go
 * details: Initializes all the message handlers
 *
 */
package msghandler

import (
	opts "github.com/Juniper/collector/flow-translator/options"
	"sync"
)

type Handler struct {
	MH     MsgHandler
	MHChan chan []byte
}

type MsgHandler interface {
	setup() error
	handleMessages(chan []byte)
}

func NewMsgHandler(handlerName string) *Handler {
	var msgHandlerRegistered = map[string]MsgHandler{
		opts.StrDataManager: new(DataManager),
		opts.StrQueryAPI:    new(QueryAPI),
	}
	return &Handler{
		MH: msgHandlerRegistered[handlerName],
	}
}

func (h Handler) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)
	err = h.MH.setup()
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.MH.handleMessages(h.MHChan)
	}()

	wg.Wait()

	return nil
}
