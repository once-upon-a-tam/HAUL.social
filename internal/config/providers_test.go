package config

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func secretStore(secrets map[string]string) ReadSecret {
	return func(key string) (string, error) {
		val, ok := secrets[key]
		if !ok {
			return "", fmt.Errorf("secret %q not found", key)
		}
		return val, nil
	}
}

func TestLoadProviders_Success(t *testing.T) {
	cases := []struct {
		name     string
		secrets  map[string]string
		expected []OAuthProvider
	}{
		{
			name: "single provider",
			secrets: map[string]string{
				"oauth_providers":      "github",
				"github_client_id":     "gh-id",
				"github_client_secret": "gh-secret",
			},
			expected: []OAuthProvider{
				{Name: "github", ClientID: "gh-id", ClientSecret: "gh-secret"},
			},
		},
		{
			name: "multiple providers",
			secrets: map[string]string{
				"oauth_providers":      "github,google",
				"github_client_id":     "gh-id",
				"github_client_secret": "gh-secret",
				"google_client_id":     "google-id",
				"google_client_secret": "google-secret",
			},
			expected: []OAuthProvider{
				{Name: "github", ClientID: "gh-id", ClientSecret: "gh-secret"},
				{Name: "google", ClientID: "google-id", ClientSecret: "google-secret"},
			},
		},
		{
			name: "provider names with surrounding whitespaces",
			secrets: map[string]string{
				"oauth_providers":      "   github ,  google   ",
				"github_client_id":     "gh-id",
				"github_client_secret": "gh-secret",
				"google_client_id":     "google-id",
				"google_client_secret": "google-secret",
			},
			expected: []OAuthProvider{
				{Name: "github", ClientID: "gh-id", ClientSecret: "gh-secret"},
				{Name: "google", ClientID: "google-id", ClientSecret: "google-secret"},
			},
		},
		{
			name: "entries between commas are skipped",
			secrets: map[string]string{
				"oauth_providers":      "github,,,    ,google",
				"github_client_id":     "gh-id",
				"github_client_secret": "gh-secret",
				"google_client_id":     "google-id",
				"google_client_secret": "google-secret",
			},
			expected: []OAuthProvider{
				{Name: "github", ClientID: "gh-id", ClientSecret: "gh-secret"},
				{Name: "google", ClientID: "google-id", ClientSecret: "google-secret"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := LoadProviders(secretStore(tc.secrets))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d providers, got %d", len(tc.expected), len(result))
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("provider[%d]: expected %v, got %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestLoadProviders_Error(t *testing.T) {
	errStore := errors.New("store unavailable")

	cases := []struct {
		name            string
		read            ReadSecret
		expectedErrWrap error
		expectedErrMsg  string
	}{
		{
			name: "read oauth_providers fails",
			read: func(key string) (string, error) {
				if key == "oauth_providers" {
					return "", errStore
				}
				return "", nil
			},
			expectedErrWrap: errStore,
			expectedErrMsg:  "failed reading the oauth_providers secret",
		},
		{
			name: "empty oauth_providers secret",
			read: secretStore(map[string]string{
				"oauth_providers": "",
			}),
			expectedErrMsg: "no valid provider names",
		},
		{
			name: "whitespace-only oauth_providers secret",
			read: secretStore(map[string]string{
				"oauth_providers": "  ,    ,  ",
			}),
			expectedErrMsg: "no valid provider names",
		},
		{
			name: "read client_id fails",
			read: func(key string) (string, error) {
				if key == "oauth_providers" {
					return "github", nil
				}
				if key == "github_client_id" {
					return "", errStore
				}
				return "", nil
			},
			expectedErrWrap: errStore,
			expectedErrMsg:  `provider "github"`,
		},
		{
			name: "read client_secret fails",
			read: func(key string) (string, error) {
				if key == "oauth_providers" {
					return "github", nil
				}
				if key == "github_client_id" {
					return "gh-id", nil
				}
				if key == "github_client_secret" {
					return "", errStore
				}
				return "", nil
			},
			expectedErrWrap: errStore,
			expectedErrMsg:  `provider "github"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := LoadProviders(tc.read)
			if err == nil {
				t.Fatalf("expected error, got providers: %v", result)
			}

			if tc.expectedErrWrap != nil && !errors.Is(err, tc.expectedErrWrap) {
				t.Errorf("error chain: expected %v, got %v", tc.expectedErrWrap, err)
			}

			if tc.expectedErrMsg != "" && !strings.Contains(err.Error(), tc.expectedErrMsg) {
				t.Errorf("error message: expected substring %q, got %q", tc.expectedErrMsg, err.Error())
			}
		})
	}
}
