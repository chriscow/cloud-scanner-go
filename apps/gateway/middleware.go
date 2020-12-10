package main

import (
	"net/http"
	"strings"
	"context"
)

func (s *server) setDefaultPageData(cfg config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			data := map[string]interface{}{
				"route":    r.URL.Path,
				"app_name": strings.Title(strings.ToLower(cfg.Name)),
			}

			_, email, metadata, err := s.appCtx.user.LoggedInUser(r)
			if err == nil {
				data["email"] = email
				data["metadata"] = metadata
				data["is_logged_in"] = true
			}

			ctx := context.WithValue(r.Context(), appCtxDataKey, data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}