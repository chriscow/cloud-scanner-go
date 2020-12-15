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

func findOEIS(appCtx appContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	var once sync.Once
	once.Do(func() {
		if err := loadOEIS(); err != nil {
			log.Fatal("load oeis database", err)
		}
	})

	r.ParseForm()

	var seq string
	var pos int
	var count int
	var err error

	input := strings.Split(r.Form["in"][0], " ")

	seq = input[0]
	pos = 0
	if len(input) > 1 {
		pos, err = strconv.Atoi(input[1])
		if err != nil {
			return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
		}
		pos = (pos - 1) * 2
	}

	tok := strings.Split(seq, "")
	seq = strings.Join(tok, ",")

	start := time.Now()
	count = 0

	type data struct {
		ID  string
		Seq string
	}

	results := make([]data, 0)

	for k, v := range oeisSeq {
		if strings.Index(v, seq) == pos {
			results = append(results, data{ID: strings.Trim(k, " "), Seq: v})
			count++
		}
	}
	elapsed := time.Since(start)
	fmt.Println("count:", count, "elapsed:", elapsed.Milliseconds())

	return goview.M{
		"elapsed": elapsed.Milliseconds(),
		"count":   count,
		"query":   r.Form["in"][0],
		"data":    results,
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
