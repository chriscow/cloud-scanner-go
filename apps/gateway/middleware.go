package main

import (
	"net/http"
	"strings"
	"context"
)

// setDefaultPageData is a middleware function that is called on every request.
// It adds a map[string]interface{} with the current route and applicaiton name
// for use in html templates.  If the user is logged in, it also adds their
// email address, provider metadata and a flag `is_logged_in`.
//
// The map is stored in the request context under the key `appCtxDataKey`
//
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