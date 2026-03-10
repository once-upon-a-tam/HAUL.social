package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(migrateCreateLinksCollection, migrateDropLinksCollection)
}

func migrateCreateLinksCollection(app core.App) error {
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	links := core.NewBaseCollection("links")
	links.ListRule = types.Pointer("user = @request.auth.id")
	links.ViewRule = types.Pointer("user = @request.auth.id")
	links.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	links.UpdateRule = types.Pointer("user = @request.auth.id")
	links.DeleteRule = types.Pointer("user = @request.auth.id")

	links.Fields.Add(
		&core.RelationField{
			Name:          "user",
			Required:      true,
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
		},
		&core.URLField{
			Name:     "url",
			Required: true,
		},
		&core.TextField{
			Name:     "note",
			Required: false,
			Max:      280,
		},
		&core.BoolField{
			Name: "stale",
		},
		&core.DateField{
			Name: "last_visited",
		},
	)

	return app.Save(links)
}

func migrateDropLinksCollection(app core.App) error {
	links, err := app.FindCollectionByNameOrId("links")
	if err != nil {
		return err
	}

	return app.Delete(links)
}
