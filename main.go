package main

import (
	authmodule "github.com/dimas292/boilerplate-rest/modules/auth"
	"github.com/dimas292/boilerplate-rest/pkg/server"
)

func main() {
	// Bootstrap server (config + postgres + redis + jwt + gin)
	srv := server.New("config.yml")

	// Register feature modules
	srv.RegisterModules(
		authmodule.NewAuthModule(srv.DB, srv.Redis, srv.JWT),
		// yourmodule.NewYourModule(srv.DB),
	)

	// Start HTTP server
	srv.Run()
}
