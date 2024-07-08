// Copyright (c) 2021 Amirhossein Movahedi (@qolzam)
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/middleware/authcookie"
	appConfig "github.com/red-gold/telar-web/micros/storage/config"
	"github.com/red-gold/telar-web/micros/storage/handlers"
)

// @title Storage micro API
// @version 1.0
// @description This is an API to handle files
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email dev@telar.dev
// @license.name MIT
// @license.url https://github.com/red-gold/telar-web/blob/master/LICENSE
// @host social.faas.telar.dev
// @BasePath /storage
// @securityDefinitions.apiKey  JWT
// @name Authorization
// @in header
func SetupRoutes(app *fiber.App) {

	if appConfig.StorageConfig.ProxyBalancer != "" {
		app.Use(proxy.Balancer(proxy.Config{
			Servers: []string{
				appConfig.StorageConfig.ProxyBalancer,
			},
		}))
	}

	// Middleware
	authCookieMiddleware := authcookie.New(authcookie.Config{
		JWTSecretKey: []byte(*config.AppConfig.PublicKey),
	})

	// Router
	app.Post("/:uid/:dir", authCookieMiddleware, handlers.UploadeHandle)
	app.Get("/:uid/:dir/:name", authCookieMiddleware, handlers.GetFileHandle)

}
