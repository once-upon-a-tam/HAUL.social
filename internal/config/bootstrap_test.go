package config

import (
	"errors"
	"testing"

	"github.com/pocketbase/pocketbase/tests"
)

func TestApplySettings_Success(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to provision test app: %v", err)
	}
	defer app.Cleanup()

	read := secretStore(map[string]string{
		"app_url": "https://example.com",
	})

	if err := ApplySettings(app, read); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings := app.Settings()

	t.Run("meta", func(t *testing.T) {
		if result, expected := settings.Meta.AppURL, "https://example.com"; result != expected {
			t.Errorf("AppURL: expected %q, got %q", expected, result)
		}
	})

	t.Run("trusted proxy", func(t *testing.T) {
		headers := settings.TrustedProxy.Headers
		if len(headers) != 1 || headers[0] != "X-Forwarded-To" {
			t.Errorf("TrustedProxy.Headers: expected %v, got %v", "[X-Forwarded-To]", headers)
		}
	})

	t.Run("rate limits", func(t *testing.T) {
		rl := settings.RateLimits
		if !rl.Enabled {
			t.Errorf("Expected RateLimits.Enabled to be true, got false")
		}

		cases := []struct {
			label       string
			maxRequests int
			duration    int64
		}{
			{"default", 120, 60},
			{"/api/", 30, 60},
		}
		if len(rl.Rules) != len(cases) {
			t.Errorf("RateLimits.Rules: expected %d rules, got %d", len(cases), len(rl.Rules))
		}

		for i, tc := range cases {
			result := rl.Rules[i]
			if result.Label != tc.label || result.MaxRequests != tc.maxRequests || result.Duration != tc.duration {
				t.Errorf("RateLimits.Rules[%d]: expected %v, got %v", i, tc, result)
			}
		}
	})
}

func TestApplySettings_Errors(t *testing.T) {
	errStore := errors.New("store unavailable")

	cases := []struct {
		name            string
		read            ReadSecret
		expectedErrWrap error
	}{
		{
			name: "read app_url fails",
			read: func(key string) (string, error) {
				return "", errStore
			},
			expectedErrWrap: errStore,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, err := tests.NewTestApp()
			if err != nil {
				t.Fatalf("failed to provision test app: %v", err)
			}
			defer app.Cleanup()

			err = ApplySettings(app, tc.read)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.expectedErrWrap) {
				t.Errorf("error chain: expected %v, got %v", tc.expectedErrWrap, err)
			}
		})
	}
}

func TestApplyOAuthProviders_Success(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to provision test app: %v", err)
	}
	defer app.Cleanup()

	read := secretStore(map[string]string{
		"oauth_providers":      "github, google",
		"github_client_id":     "gh-id",
		"github_client_secret": "gh-secret",
		"google_client_id":     "google-id",
		"google_client_secret": "google-secret",
	})

	if err := ApplyOAuthProviders(app, read); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("could not fetch users collection: %v", err)
	}

	t.Run("auth configuration", func(t *testing.T) {
		cases := []struct {
			name     string
			expected bool
			result   bool
		}{
			{"PasswordAuth.Enabled", users.PasswordAuth.Enabled, false},
			{"OAuth2.Enabled", users.OAuth2.Enabled, true},
		}

		for _, tc := range cases {
			if tc.result != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, tc.result)
			}
		}
	})

	t.Run("OAuth2 providers", func(t *testing.T) {
		expected := []struct {
			name         string
			clientID     string
			clientSecret string
		}{
			{"github", "gh-id", "gh-secret"},
			{"google", "google-id", "google-secret"},
		}

		if len(users.OAuth2.Providers) != len(expected) {
			t.Fatalf("Expected %d OAuth2 providers, got %d", len(expected), len(users.OAuth2.Providers))
		}

		for i, provider := range expected {
			result := users.OAuth2.Providers[i]
			if result.Name != provider.name || result.ClientId != provider.clientID || result.ClientSecret != provider.clientSecret {
				t.Errorf("users.OAuth2.Providers[%d]: expected %v, got %v", i, provider, result)
			}
		}
	})
}

func TestApplyOAuthProviders_Error(t *testing.T) {
	errStore := errors.New("store unavailable")

	cases := []struct {
		name            string
		read            ReadSecret
		expectedErrWrap error
	}{
		{
			name: "LoadProviders fails",
			read: func(key string) (string, error) {
				return "", errStore
			},
			expectedErrWrap: errStore,
		},
		{
			name: "empty providers list",
			read: secretStore(map[string]string{"oauth_providers": ""}),
		},
	}

	for _, tc := range cases {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatalf("failed to provision test app: %v", err)
		}
		defer app.Cleanup()

		err = ApplyOAuthProviders(app, tc.read)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if tc.expectedErrWrap != nil && !errors.Is(err, tc.expectedErrWrap) {
			t.Errorf("error chain: expected %v, got %v", tc.expectedErrWrap, err)
		}
	}
}
