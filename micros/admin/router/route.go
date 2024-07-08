// Copyright (c) 2021 Amirhossein Movahedi (@qolzam)
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/middleware/authcookie"
	"github.com/red-gold/telar-core/middleware/authrole"
	"github.com/red-gold/telar-web/micros/admin/handlers"
)

// @title Admin micro API
// @version 1.0
// @description This is an API to handle admin operations
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email dev@telar.dev
// @license.name MIT
// @license.url https://github.com/red-gold/telar-web/blob/master/LICENSE
// @host social.faas.telar.dev
// @BasePath /admin
func SetupRoutes(app *fiber.App) {

	// Middleware
	authCookieMiddleware := authcookie.New(authcookie.Config{
		JWTSecretKey: []byte(*config.AppConfig.PublicKey),
	})
	authRoleMiddleware := authrole.New(authrole.ConfigDefault)

	// Router
	app.Post("/setup", authCookieMiddleware, authRoleMiddleware, handlers.SetupHandler)
	app.Get("/setup", authCookieMiddleware, authRoleMiddleware, handlers.SetupPageHandler)
	app.Get("/login", handlers.LoginPageHandler)
	app.Post("/login", handlers.LoginAdminHandler)
}
