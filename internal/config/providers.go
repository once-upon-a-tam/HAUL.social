package config

import (
	"errors"
	"fmt"
	"strings"
)

type OAuthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string // #nosec G117 // Will never get sent over API
}

type ReadSecret func(key string) (string, error)

func LoadProviders(read ReadSecret) ([]OAuthProvider, error) {
	raw, err := read("oauth_providers")
	if err != nil {
		return nil, fmt.Errorf("failed reading the oauth_providers secret: %w", err)
	}

	var providers []OAuthProvider
	for name := range strings.SplitSeq(raw, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		clientID, err := read(name + "_client_id")
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", name, err)
		}

		clientSecret, err := read(name + "_client_secret")
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", name, err)
		}

		providers = append(providers, OAuthProvider{
			Name:         name,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		})
	}

	if len(providers) == 0 {
		return nil, errors.New("oauth_providers secret contains no valid provider names")
	}

	return providers, nil
}
