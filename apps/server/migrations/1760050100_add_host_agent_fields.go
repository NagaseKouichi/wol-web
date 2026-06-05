package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("hosts")
		if err != nil {
			return err
		}

		if collection.Fields.GetByName("agentUrl") == nil {
			collection.Fields.Add(&core.TextField{
				Name:     "agentUrl",
				Required: false,
			})
		}

		if collection.Fields.GetByName("agentToken") == nil {
			collection.Fields.Add(&core.TextField{
				Name:     "agentToken",
				Hidden:   true,
				Required: false,
			})
		} else if field, ok := collection.Fields.GetByName("agentToken").(*core.TextField); ok {
			field.Hidden = true
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("hosts")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("agentUrl")
		collection.Fields.RemoveByName("agentToken")

		return app.Save(collection)
	})
}
