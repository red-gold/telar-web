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
	"github.com/red-gold/telar-web/micros/profile/handlers"
)

// SetupRoutes func
func SetupRoutes(app *fiber.App) {

	// Middleware
	authHMACMiddleware := func(hmacWithCookie bool) func(*fiber.Ctx) error {
		var Next func(c *fiber.Ctx) bool = nil
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
		var Next func(c *fiber.Ctx) bool = nil
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

	// Routers
	app.Get("/my", authCookieMiddleware(false), handlers.ReadMyProfileHandle)
	app.Get("/", append(hmacCookieHandlers, handlers.QueryUserProfileHandle)...)
	app.Get("/id/:userId", append(hmacCookieHandlers, handlers.ReadProfileHandle)...)
	app.Post("/index", authHMACMiddleware(false), handlers.InitProfileIndexHandle)
	app.Put("/last-seen", authHMACMiddleware(false), handlers.UpdateLastSeen)

	// Invoke between functions and protected by HMAC
	app.Put("/", authHMACMiddleware(false), handlers.UpdateProfileHandle)
	app.Get("/dto/id/:userId", authHMACMiddleware(false), handlers.ReadDtoProfileHandle)
	app.Post("/dto", authHMACMiddleware(false), handlers.CreateDtoProfileHandle)
	app.Post("/dispatch", authHMACMiddleware(false), handlers.DispatchProfilesHandle)
	app.Post("/dto/ids", authHMACMiddleware(false), handlers.GetProfileByIds)
	app.Put("/follow/inc/:inc/:userId", authHMACMiddleware(false), handlers.IncreaseFollowCount)
	app.Put("/follower/inc/:inc/:userId", authHMACMiddleware(false), handlers.IncreaseFollowerCount)
}
