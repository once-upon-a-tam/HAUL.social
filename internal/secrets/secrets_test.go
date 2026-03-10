package secrets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSecretFile(t *testing.T, dir, key, value string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, key), []byte(value), 0o600); err != nil {
		t.Fatalf("writeSecretFile: %v", err)
	}
}

func unsetenv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("os.Unsetenv(%q): %v", key, err)
	}
}

func TestRead_FromFile_ExactValue(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "my_key", "supersecretvalue")

	result, err := read(dir, "my_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "supersecretvalue" {
		t.Errorf("got %q, expected %q", result, "supersecretvalue")
	}
}

func TestRead_FromFile_TrimWhitespace(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "trailing newline", value: "supersecretvalue\n", expected: "supersecretvalue"},
		{name: "trailing spaces", value: "supersecretvalue", expected: "supersecretvalue"},
		{name: "leading and trailing whitespace", value: "   supersecretvalue\n", expected: "supersecretvalue"},
		{name: "only whitespace, returns empty string", value: "      \n", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeSecretFile(t, dir, "my_key", tc.value)
			result, err := read(dir, "my_key")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Fatalf("value=%q: got %q, expected %q", tc.value, result, tc.expected)
			}
		})
	}
}

func TestRead_FromFile_TakesPrecedenceOverEnv(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "my_key", "from_file")
	t.Setenv("MY-KEY", "from_env")

	result, err := read(dir, "my_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "from_env" {
		t.Fatalf("expected %q but got %q, indicating the env var took precedence over the secrets file", "from_file", result)
	}

	if result != "from_file" {
		t.Fatalf("expected %q, but got %q", "from_file", result)
	}
}

func TestRead_FromEnv(t *testing.T) {
	cases := []struct {
		name     string
		envKey   string
		readKey  string
		value    string
		expected string
	}{
		{name: "exact match", envKey: "MY_SECRET_KEY", readKey: "MY_SECRET_KEY", value: "envvalue", expected: "envvalue"},
		{name: "key converted to upper case", envKey: "MY_SECRET_KEY", readKey: "my_secret_key", value: "uppercased", expected: "uppercased"},
		{name: "key case normalised", envKey: "MY_SECRET_KEY", readKey: "My_Secret_Key", value: "envvalue", expected: "envvalue"},
		// Values are only trimmed for file content, env values are used verbatim
		{name: "non trimmed valued", envKey: "MY_SECRET_KEY", readKey: "MY_SECRET_KEY", value: "  envvalue  ", expected: "  envvalue  "},
	}

	for _, tc := range cases {
		dir := t.TempDir()
		t.Setenv(tc.envKey, tc.value)

		result, err := read(dir, tc.readKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != tc.expected {
			t.Fatalf("expected %q, got %q", tc.expected, result)
		}
	}
}

func TestRead_NotFound(t *testing.T) {
	cases := []struct {
		name    string
		envKey  string
		readKey string
		checkFn func(t *testing.T, val string, err error)
	}{
		{
			name:    "returns non-nil error",
			envKey:  "MISSING_KEY_XYZ",
			readKey: "missing_key_xyz",
			checkFn: func(t *testing.T, _ string, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("expected an error, but go nil")
				}
			},
		},
		{
			name:    "error message contains the key",
			envKey:  "MISSING_KEY_ABC",
			readKey: "missing_key_abc",
			checkFn: func(t *testing.T, _ string, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("expected an error, but go nil")
				}
				msg := err.Error()
				if !strings.Contains(msg, "missing_key_abc") {
					t.Fatalf("error %q doesn't mention the missing key's name (%q)", msg, "missing_key_abc")
				}
			},
		},
		{
			name:    "error message mentions both sources",
			envKey:  "MISSING_BOTH_42",
			readKey: "missing_both_42",
			checkFn: func(t *testing.T, _ string, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("expected an error, but go nil")
				}
				msg := err.Error()
				if !strings.Contains(msg, "environment") {
					t.Fatalf("error %q doesn't mention environment variables", msg)
				}
			},
		},
		{
			name:    "returns empty string alongside error",
			envKey:  "EMPTY_RETURN_KEY",
			readKey: "empty_return_key",
			checkFn: func(t *testing.T, result string, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("expected an error, but go nil")
				}
				if result != "" {
					t.Fatalf("expected empty string result on failure, got %q", result)
				}
			},
		},
		// os.GetEnv returns "" for unset variables. the read function treats that
		// as an absent variable.
		{
			name:    "empty-string env var falls through to error",
			envKey:  "EMPTY_VALUE_KEY",
			readKey: "empty_value_key",
			checkFn: func(t *testing.T, _ string, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("expected an error for an empty-string env variable, but got nil")
				}
			},
		},
	}

	for _, tc := range cases {
		dir := t.TempDir()
		unsetenv(t, tc.envKey)

		result, err := read(dir, tc.readKey)
		tc.checkFn(t, result, err)
	}
}
