package main

import (
	"haul/internal/config"
	"haul/internal/secrets"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	app.OnBootstrap().BindFunc(func(be *core.BootstrapEvent) error {
		if err := be.Next(); err != nil {
			return err
		}
		if err := config.ApplySettings(app, secrets.Read); err != nil {
			return err
		}
		if err := config.ApplyOAuthProviders(app, secrets.Read); err != nil {
			return err
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
