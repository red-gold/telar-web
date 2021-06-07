// Copyright (c) 2021 Amirhossein Movahedi (@qolzam)
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/middleware/authcookie"
	"github.com/red-gold/telar-core/middleware/authhmac"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-web/micros/notifications/handlers"
)

// SetupRoutes func
func SetupRoutes(app *fiber.App) {

	// Middleware
	authHMACMiddleware := func(hmacWithCookie bool) func(*fiber.Ctx) error {
		var Next func(c *fiber.Ctx) bool
		if hmacWithCookie {
			Next = func(c *fiber.Ctx) bool {
				if c.Get(types.HeaderHMACAuthenticate) != "" {
					return false
				}
				return true
			}
		}
		return authhmac.New(authhmac.Config{
			Next:          Next,
			PayloadSecret: *config.AppConfig.PayloadSecret,
		})
	}

	authCookieMiddleware := func(hmacWithCookie bool) func(*fiber.Ctx) error {
		var Next func(c *fiber.Ctx) bool
		if hmacWithCookie {
			Next = func(c *fiber.Ctx) bool {
				if c.Get(types.HeaderHMACAuthenticate) != "" {
					return true
				}
				return false
			}
		}
		return authcookie.New(authcookie.Config{
			Next:         Next,
			JWTSecretKey: []byte(*config.AppConfig.PublicKey),
		})
	}

	hmacCookieHandlers := []func(*fiber.Ctx) error{authHMACMiddleware(true), authCookieMiddleware(true)}

	// Router
	app.Post("/check", authHMACMiddleware(false), handlers.CheckNotifyEmailHandle)
	app.Post("/", authHMACMiddleware(false), handlers.CreateNotificationHandle)
	app.Put("/", authHMACMiddleware(false), handlers.UpdateNotificationHandle)
	app.Put("/seen/:notificationId", append(hmacCookieHandlers, handlers.SeenNotificationHandle)...)
	app.Delete("/id/:notificationId", append(hmacCookieHandlers, handlers.DeleteNotificationHandle)...)
	app.Delete("/my", append(hmacCookieHandlers, handlers.DeleteNotificationByUserIdHandle)...)
	app.Get("/", append(hmacCookieHandlers, handlers.GetNotificationsByUserIdHandle)...)
	app.Get("/:notificationId", append(hmacCookieHandlers, handlers.GetNotificationHandle)...)
}
