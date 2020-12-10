package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/foolin/goview"
)

type pageHandlerFunc func(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error)

func loginPage(_ appContext, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
	confirmed := r.URL.Query().Get("confirmed")
	if confirmed == "true" {
		return goview.M{
			"confirmed": true,
		}, nil
	}

	notConfirmed := r.URL.Query().Get("not_confirmed")
	if notConfirmed == "true" {
		return goview.M{
			"not_confirmed": true,
		}, nil
	}

	confirmationSent := r.URL.Query().Get("confirmation_sent")
	if confirmationSent == "true" {
		return goview.M{
			"confirmation_sent": true,
		}, nil
	}

	emailChanged := r.URL.Query().Get("email_changed")
	if emailChanged == "true" {
		return goview.M{
			"email_changed": true,
		}, nil
	}

	return goview.M{}, nil
}


func authPage(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	err := appCtx.user.HandleGothLogin(w, r)
	if err != nil {
		return goview.M{}, err
	}
	redirectTo := "/app"
	from := r.URL.Query().Get("from")
	if from != "" {
		redirectTo = from
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return goview.M{}, nil
}


func authCallbackPage(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	err := appCtx.user.HandleGothCallback(w, r)
	if err != nil {
		return goview.M{}, err
	}
	redirectTo := "/app"
	from := r.URL.Query().Get("from")
	if from != "" {
		redirectTo = from
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return goview.M{}, nil
}

func appPage(_ appContext, _ http.ResponseWriter, _ *http.Request) (goview.M, error) {
	dummy := struct {
		Title string `json:"title"`
	}{
		Title: "Hello Props",
	}

	d, err := json.Marshal(&dummy)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
	}

	return goview.M{
		"Data": string(d),
	}, nil
}