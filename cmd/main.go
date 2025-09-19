package main

import (
	"cmsapp/devmode"
	"cmsapp/internal/db"
	"github.com/gofiber/fiber/v2"
	"log"
	"cmsapp/internal/user"
	"cmsapp/internal/auth"
	"cmsapp/internal/media"
	"cmsapp/internal/content"
)	

func main() {

	setting.LoadEnv()

	//Aws
	media.InitS3()
	if err := media.InitS3(); err != nil {
    log.Fatalf("failed to init S3: %v", err)
}

	//Database
	db.Connect()
	db.DB.AutoMigrate(
		&user.User{},
		&media.MediaAsset{},
    &content.ContentModel{},
    &content.ContentEntry{},
	&content.EntryRelation{},
	)	
	//Route
	app:= fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "CMS API running 🚀",
		})
	})
	//User
	app.Post("/register", user.RegisterHandler)
	app.Post("/login", user.LoginHandler)
	app.Get("/me", auth.AuthRequired, user.MeHandler)
	
	// Models
	app.Post("/models", auth.AuthRequired, auth.RoleRequired("admin"), content.CreateModelHandler)
	app.Put("/models/:id", auth.AuthRequired, auth.RoleRequired("admin"), content.UpdateModelHandler)
	app.Delete("/models/:id", auth.AuthRequired, auth.RoleRequired("admin"), content.DeleteModelHandler)
	app.Get("/models", auth.AuthRequired, auth.RoleRequired("admin", "editor", "viewer"), content.ListModelsHandler)

	// Entries (admin/editor/viewer)
	app.Post("/entries", auth.AuthRequired, auth.RoleRequired("admin", "editor"), content.CreateEntryHandler)
	app.Put("/entries/:id", auth.AuthRequired, auth.RoleRequired("admin", "editor"), content.UpdateEntryHandler)
	app.Delete("/entries/:id", auth.AuthRequired, auth.RoleRequired("admin"), content.DeleteEntryHandler)
	app.Get("/entries/:id", auth.AuthRequired, auth.RoleRequired("admin", "editor", "viewer"), content.GetEntryHandler)
	app.Get("/entries/:model_id", auth.AuthRequired, auth.RoleRequired("admin", "editor", "viewer"), content.ListEntriesHandler)
	app.Post("/entries/:id/:action", auth.AuthRequired, auth.RoleRequired("admin"), content.SetEntryStatusHandler)
	app.Get("/entries/:id/audit", auth.AuthRequired, auth.RoleRequired("admin"), content.GetAuditLogsHandler)


	// Public (no auth)
	app.Get("/public/:model/:slug", content.GetPublicEntryHandler)
	app.Get("/public/:model", content.ListPublicEntriesHandler)
	app.Get("/public/:model/:slug", content.GetPublicEntryHandler)

	// Media (upload only editor/admin)
	app.Get("/media/presign", auth.AuthRequired, auth.RoleRequired("admin", "editor"), media.PresignHandler)
	app.Get("/media", auth.AuthRequired, auth.RoleRequired("admin", "editor", "viewer"), media.GetMedia)
	app.Delete("/media/:id", auth.AuthRequired, auth.RoleRequired("admin"), media.DeleteMedia)

	app.Listen(":3000")
}