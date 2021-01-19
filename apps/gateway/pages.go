package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

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

var once sync.Once

func findOEIS(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	once.Do(func() {
		if err := loadOEIS(); err != nil {
			log.Fatal("load oeis sequences", err)
		}

		if err := loadOEISDecExp(); err != nil {
			log.Fatal("load oeis names", err)
		}
	})

	in := r.URL.Query().Get("in")
	if in == "" {
		return goview.M{}, nil
	}

	deOnly := false
	if r.URL.Query().Get("de") != "" {
		deOnly = true
	}

	var err error
	pos := 0

	// the input sequence is a string of numbers and optionally a space
	// to indicate starting position in the list
	if strings.Contains(in, " ") {
		tok := strings.Split(in, " ")
		in = tok[0]

		// see if a position was entered
		if len(tok) > 1 {
			pos, err = strconv.Atoi(tok[1])
			if err != nil {
				return goview.M{"error": err.Error()}, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
			}

			if pos > 0 {
				pos-- // Input is 1-based so change to zero-based
			}
		}
	}

	seq := strings.Split(in, "")
	digits := make([]int, len(seq))

	for i := range seq {
		dig, err := strconv.Atoi(seq[i])
		if err != nil {
			return goview.M{
				"error": err.Error(),
			}, err
		}

		digits[i] = dig
	}

	start := time.Now()
	results, err := searchOEIS(digits, pos, deOnly)
	if err != nil {
		return goview.M{"error": err.Error()}, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
	}

	count := len(results)
	fmt.Println("count:", count, "elapsed:", time.Since(start))

	return goview.M{
		"elapsed":           elapsed.Milliseconds(),
		"count":             count,
		"pos":               pos,
		"digits":            digits,
		"in":                r.URL.Query().Get("in"),
		"decimal_expansion": r.URL.Query().Get("de"),
		"data":              results,
	}, nil
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
