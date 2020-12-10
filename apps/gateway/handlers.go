// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"github.com/go-chi/chi"
	"github.com/foolin/goview"
)

type pageHandlerFunc func(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error)

// handleSubscribe handles websocket requests to subscribe to an NSQ topic
func (s *server) handleSubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		topic := chi.URLParam(r, "topic") 
		if topic == "" {
			http.Error(w, "Invalid request", http.StatusNotAcceptable)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Websocket error:"+err.Error(), http.StatusInternalServerError)
			return
		}

		
		pub := s.getPublication(topic)
		sub := newSubscriber(pub, conn)
		pub.subscribe <- sub
	}
}