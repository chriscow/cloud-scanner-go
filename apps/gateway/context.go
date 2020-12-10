package main

import (
	"github.com/foolin/goview"
	"github.com/adnaan/users"
)

type appContext struct {
	user       *users.API
	viewEngine *goview.ViewEngine
	pageData   goview.M
	cfg        config
}