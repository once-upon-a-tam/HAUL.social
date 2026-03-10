package config

import (
	"github.com/pocketbase/pocketbase/core"
)

func ApplySettings(app core.App, read ReadSecret) error {	
	appURL, err := read("app_url")
	if err != nil {
		return err
	}

	settings := app.Settings()
	settings.Meta.AppURL = appURL
	settings.TrustedProxy.Headers = []string{"X-Forwarded-To"}
	settings.RateLimits.Enabled = true
	settings.RateLimits.Rules = []core.RateLimitRule{
		{Label: "default", MaxRequests: 120, Duration: 60},
		{Label: "/api/", MaxRequests: 30, Duration: 60},
	}

	return app.Save(settings)
}

func ApplyOAuthProviders(app core.App, read ReadSecret) error {
	providers, err := LoadProviders(read)
	if err != nil {
		return err
	}

	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	// TODO: Change these lines to a flag once we allow basic auth usage
	users.PasswordAuth.Enabled = false
	users.OAuth2.Enabled = true

	users.OAuth2.Providers = make([]core.OAuth2ProviderConfig, len(providers))

	for i, p := range providers {
		users.OAuth2.Providers[i] = core.OAuth2ProviderConfig{
			Name: p.Name,
			ClientId: p.ClientID,
			ClientSecret: p.ClientSecret,
		}
	}

	return app.Save(users)
}
