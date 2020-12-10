package main

import (
	"github.com/foolin/goview"
	"time"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/valve"
	"github.com/Masterminds/sprig"
	"github.com/adnaan/users"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)
const appCtxDataKey = "app_ctx_data"


type server struct {
	cfg 			config	
	appCtx 			appContext		
	router 			chi.Router
	publications	map[string]*publication
	valve			*valve.Valve
	view 			*goview.ViewEngine
}

func newServer(cfg config) *server {
	viewCfg := goview.DefaultConfig
	viewCfg.DisableCache = true
	viewCfg.Funcs = sprig.FuncMap()

	s := &server{
		cfg: 			cfg,
		router: 		chi.NewRouter(),
		publications: 	make(map[string]*publication),
		valve:     		valve.New(),
		view: 			goview.New(viewCfg),
	}

	s.configure()

	return s
}


func (s *server) run(addr string) (err error) {
	fmt.Println("Listening on ", addr)
	http.ListenAndServe(addr, s.router)
	return s.valve.Shutdown(20 * time.Second)
}


func (s *server) configure() {
	s.context()
	s.middleware()
	s.routes()
}

func (s *server) context() error {
	
	defaultUsersConfig := users.Config{
		Driver:        s.cfg.Driver,
		Datasource:    s.cfg.DataSource,
		SessionSecret: s.cfg.SessionSecret,
		GothProviders: []goth.Provider{
			google.New(s.cfg.GoogleClientID, s.cfg.GoogleSecret, fmt.Sprintf("%s/auth/callback?provider=google", s.cfg.Domain), "email", "profile"),
		},
	}
	usersAPI, err := users.NewDefaultAPI(s.valve.Context(), defaultUsersConfig)
	if err != nil {
		return err
	}

	s.appCtx = appContext{
		user:		usersAPI,
		viewEngine: s.view,
		pageData:   goview.M{},
		cfg:        s.cfg,
	}

	return nil
}

func (s *server) middleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Heartbeat("/ping"))
	// s.router.Use(render.SetContentType(render.ContentTypeJSON))
	s.router.Use(middleware.Compress(5)) // be sure to set w.Header().Set("Content-Type", http.DetectContentType(yourBody))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)
	s.router.Use(s.setDefaultPageData(s.cfg))
}

func (s *server) routes() {
	// static file route
	s.staticRoute("/static")

	// returns a function that takes a page and one or more pageHandler funcs
	render := newRenderer(s.appCtx)

	s.router.Get("/", render("home"))
	s.router.Get("/login", render("login", loginPage) )
	s.router.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		s.appCtx.user.HandleGothLogout(w, r)
	})

	s.router.Get("/auth/callback", render("login", authCallbackPage))
	s.router.Get("/auth", render("login", authPage))

	s.router.Route("/app", func(r chi.Router) {
		r.Use(s.appCtx.user.IsAuthenticated)
		r.Get("/", render("app", appPage))
	})

	s.router.Route("/api", func(r chi.Router) {
		r.Use(s.appCtx.user.IsAuthenticated)
		r.Use(middleware.AllowContentType("application/json"))

		r.Route("/session", func(r chi.Router) {
			// get a session by it's ID from the database or return a "default" session
			r.Get("/", getDefaultSession)
			r.Get("/{sessionID}", getSession)
	
			// queue a scan using the parameters of the session
			r.Post("/", startSession)
		})

	})
	
	s.router.Get("/subscribe/{topic}", s.handleSubscribe())	
}

// sets up a http.FileServer handler to serve static files from a http.FileSystem
func (s *server) staticRoute(staticPath string) {
	if strings.ContainsAny(staticPath, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	// location of public static files on the server
	cwd, _ := os.Getwd()
	public := http.Dir(path.Join(cwd, "./public"))

	// set up a redirect if the path does not end in a slash to one that does
	if staticPath != "/" && staticPath[len(staticPath)-1] != '/' {
		s.router.Get(staticPath, http.RedirectHandler(staticPath+"/", 301).ServeHTTP)
		staticPath += "/"
	}
	staticPath += "*"

	s.router.Get(staticPath, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(public))
		fs.ServeHTTP(w, r)
	})
}


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

func (s *server) getPublication(topic string) *publication {
	var pub *publication
	if pub, ok := s.publications[topic]; !ok {
		pub = newPublication(s.valve.Context(), topic)
		s.publications[topic] = pub
	} 

	return pub
}