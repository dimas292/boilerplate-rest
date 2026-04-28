package main

import (
	tagmodule "github.com/dimas292/url_shortener/modules/tag"
	urlmodule "github.com/dimas292/url_shortener/modules/url"
	"github.com/dimas292/url_shortener/pkg/server"
)

func main() {
	// Bootstrap server (config + postgres + redis + gin)
	srv := server.New("config.yml")

	// Register feature modules
	srv.RegisterModules(
		urlmodule.NewURLModule(srv.DB),  // CRUD + custom endpoints
		tagmodule.NewTagModule(srv.DB),  // Pure CRUD — zero boilerplate
	)

	// Start HTTP server
	srv.Run()
}
