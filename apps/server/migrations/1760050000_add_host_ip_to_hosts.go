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

		if collection.Fields.GetByName("hostIp") != nil {
			return nil
		}

		collection.Fields.Add(&core.TextField{
			Name:     "hostIp",
			Required: false,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("hosts")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("hostIp")

		return app.Save(collection)
	})
}
