package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	// "github.com/Masterminds/sprig"
	"github.com/foolin/goview"
)

func first(str string) string {
	if len(str) == 0 {
		return ""
	}
	tmp := []rune(str)
	tmp[0] = unicode.ToUpper(tmp[0])
	return string(tmp)
}

func simplePage(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	return goview.M{}, nil
}

func newRenderer(appCtx appContext) func(page string, pageHandlerFuncs ...pageHandlerFunc) http.HandlerFunc {
	return func(page string, pageHandlerFuncs ...pageHandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// merge app context data set by isAuthenticated Middleware with the passed page data
			pageData := make(map[string]interface{})
			// default page data set by setPageData middleware
			appCtxData, ok := r.Context().Value(appCtxDataKey).(map[string]interface{})
			if ok {
				for k, v := range appCtxData {
					pageData[k] = v
				}
			}
			// set default page renderer
			if len(pageHandlerFuncs) == 0 {
				pageHandlerFuncs = append(pageHandlerFuncs, simplePage)
			}

			for _, pageHandlerFunc := range pageHandlerFuncs {
				appCtx.pageData = pageData
				pageHandlerData, err := pageHandlerFunc(appCtx, w, r)
				if err != nil {
					fmt.Println(err)
					userError := errors.Unwrap(err)
					if userError != nil {
						pageData["userError"] = first(strings.ToLower(userError.Error()))
					} else {
						pageData["userError"] = "Internal Error"
					}
				}
				// set returned page data from the handler to the main pageData map
				for k, v := range pageHandlerData {
					pageData[k] = v
				}
			}

			err := appCtx.viewEngine.Render(w, http.StatusOK, page, pageData)
			if err != nil {
				fmt.Println("viewEngine.Render error: ", err)
				fmt.Fprintf(w, "umm...awkward.")
				return
			}
		})
	}
}
