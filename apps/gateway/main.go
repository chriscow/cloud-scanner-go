//
// REST
// ====
// This example demonstrates a HTTP REST web service with some fixture data.
// Follow along the example and patterns.
//
// Also check routes.json for the generated docs from passing the -routes flag
//
// Boot the server:
// ----------------
// $ go run main.go
//
// Client requests:
// ----------------
// $ curl http://localhost:3333/
// root.
//
// $ curl http://localhost:3333/articles
// [{"id":"1","title":"Hi"},{"id":"2","title":"sup"}]
//
// $ curl http://localhost:3333/articles/1
// {"id":"1","title":"Hi"}
//
// $ curl -X DELETE http://localhost:3333/articles/1
// {"id":"1","title":"Hi"}
//
// $ curl http://localhost:3333/articles/1
// "Not Found"
//
// $ curl -X POST -d '{"id":"will-be-omitted","title":"awesomeness"}' http://localhost:3333/articles
// {"id":"97","title":"awesomeness"}
//
// $ curl http://localhost:3333/articles/97
// {"id":"97","title":"awesomeness"}
//
// $ curl http://localhost:3333/articles
// [{"id":"2","title":"sup"},{"id":"97","title":"awesomeness"}]
//
package main

import (
	"flag"
	"io"
	"os"
	"log"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, out io.Writer) error {
	flags  := flag.NewFlagSet(args[0], flag.ExitOnError)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	configFile := flag.String("config", "", "path to config file")
	envPrefix := os.Getenv("ENV_PREFIX")
	if envPrefix == "" {
		envPrefix = "app"
	}

	cfg, err := loadConfig(*configFile, envPrefix)
	if err != nil {
		log.Fatal(err)
	}

	addr   := flag.String("addr", ":4000", "http service address")
	// routes := flag.Bool("routes", false, "Generate router documentation")
	
	server := newServer(cfg)
	return server.run(*addr)
}
