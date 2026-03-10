package migrations

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func newTestApp(t testing.TB) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to provision test app: %v", err)
	}

	if err := migrateDropLinksCollection(app); err != nil {
		app.Cleanup()
		t.Fatalf("failed to revert migration state: %v", err)
	}

	return app
}

func TestMigrationLinks_Up(t *testing.T) {
	app := newTestApp(t)
	defer app.Cleanup()

	if err := migrateCreateLinksCollection(app); err != nil {
		t.Fatalf("up migration failed unexpectedly: %v", err)
	}

	links, err := app.FindCollectionByNameOrId("links")
	if err != nil {
		t.Fatalf("links collection not found after migration: %v", err)
	}

	t.Run("access rules", func(t *testing.T) {
		rules := map[string]*string{
			"ListRule":   links.ListRule,
			"ViewRule":   links.ViewRule,
			"UpdateRule": links.UpdateRule,
			"DeleteRule": links.DeleteRule,
		}

		for name, rule := range rules {
			if rule == nil {
				t.Errorf("%s: expected a rule, but got nil", name)
				continue
			}
			if expected := "user = @request.auth.id"; *rule != expected {
				t.Errorf("%s: expected %q, got %q", name, expected, *rule)
			}
		}

		if links.CreateRule == nil {
			t.Errorf("%s: expected a rule, but got nil", "CreateRule")
		} else {
			expected := "@request.auth.id != '' && @request.body.user = @request.auth.id"
			if *links.CreateRule != expected {
				t.Errorf("%s: expected %q, got %q", "CreateRule", expected, *links.CreateRule)
			}
		}
	})

	t.Run("fields", func(t *testing.T) {
		cases := []struct {
			name      string
			fieldType string
		}{
			{"user", core.FieldTypeRelation},
			{"url", core.FieldTypeURL},
			{"note", core.FieldTypeText},
			{"stale", core.FieldTypeBool},
			{"last_visited", core.FieldTypeDate},
		}

		for _, tc := range cases {
			field := links.Fields.GetByName(tc.name)
			if field == nil {
				t.Errorf("field %q not found", tc.name)
				continue
			}
			if field.Type() != tc.fieldType {
				t.Errorf("field %q has type %q, expected %q", tc.name, field.Type(), tc.fieldType)
			}
		}
	})

	t.Run("user field configuration", func(t *testing.T) {
		field, ok := links.Fields.GetByName("user").(*core.RelationField)
		if !ok {
			t.Errorf("user field: expected type RelationField, got %v", field.Type())
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			t.Errorf("could not find the users collection: %v", err)
		}

		cases := []struct {
			name     string
			result   any
			expected any
		}{
			{"Required", field.Required, true},
			{"CollectionId", field.CollectionId, users.Id},
			{"CascadeDelete", field.CascadeDelete, true},
			{"MaxSelect", field.MaxSelect, 1},
		}

		for _, tc := range cases {
			if tc.expected != tc.result {
				t.Errorf("user field: expected %s to be %v, got %v", tc.name, tc.expected, tc.result)
			}
		}
	})

	t.Run("url field configuration", func(t *testing.T) {
		field, ok := links.Fields.GetByName("url").(*core.URLField)
		if !ok {
			t.Errorf("url: expected type TextField, got %v", field.Type())
		}

		cases := []struct {
			name     string
			result   any
			expected any
		}{
			{"Required", field.Required, true},
		}

		for _, tc := range cases {
			if tc.expected != tc.result {
				t.Errorf("url field: expected %s to be %v, got %v", tc.name, tc.expected, tc.result)
			}
		}
	})

	t.Run("note field configuration", func(t *testing.T) {
		field, ok := links.Fields.GetByName("note").(*core.TextField)
		if !ok {
			t.Errorf("note: expected type TextField, got %v", field.Type())
		}

		cases := []struct {
			name     string
			result   any
			expected any
		}{
			{"Required", field.Required, false},
			{"Max", field.Max, 280},
		}

		for _, tc := range cases {
			if tc.expected != tc.result {
				t.Errorf("note field: expected %s to be %v, got %v", tc.name, tc.expected, tc.result)
			}
		}
	})
}

func TestMigrationLinks_Down(t *testing.T) {
	app := newTestApp(t)
	defer app.Cleanup()

	// Apply the up migration first, so there's something to roll back
	if err := migrateCreateLinksCollection(app); err != nil {
		t.Fatalf("prerequisite up migration failed: %v", err)
	}

	if err := migrateDropLinksCollection(app); err != nil {
		t.Fatalf("down migration failed: %v", err)
	}

	if _, err := app.FindCollectionByNameOrId("links"); err == nil {
		t.Error("links collection still exists after down migration")
	}
}

func TestMigrationLinks_UpIdempotent(t *testing.T) {
	app := newTestApp(t)
	defer app.Cleanup()

	if err := migrateCreateLinksCollection(app); err != nil {
		t.Fatalf("first up migration failed: %v", err)
	}
	// Running the up migration a second time must throw an error, not silently fail
	if err := migrateCreateLinksCollection(app); err == nil {
		t.Error("expected error when running up migration twice, got nil")
	}
}

func TestMigrationLinks_DownWithoutUp(t *testing.T) {
	app := newTestApp(t)
	defer app.Cleanup()

	if err := migrateDropLinksCollection(app); err == nil {
		t.Error("expected error when rolling back non-existent collection, got nil")
	}
}
