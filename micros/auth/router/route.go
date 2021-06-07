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
	"github.com/red-gold/telar-web/micros/auth/handlers"
)

// SetupRoutes func
func SetupRoutes(app *fiber.App) {

	// Middleware
	authHMACMiddleware := authhmac.New(authhmac.Config{
		PayloadSecret: *config.AppConfig.PayloadSecret,
	})

	authCookieMiddleware := authcookie.New(authcookie.Config{
		JWTSecretKey: []byte(*config.AppConfig.PublicKey),
	})
	admin := app.Group("/admin", authHMACMiddleware)
	login := app.Group("/login")

	//Admin
	admin.Post("/check", handlers.CheckAdminHandler)
	admin.Post("/signup", handlers.AdminSignupHandle)
	admin.Post("/login", handlers.LoginAdminHandler)

	// Signup
	app.Post("/signup/verify", handlers.VerifySignupHandle)
	app.Post("/signup", handlers.SignupTokenHandle)
	app.Get("/signup", handlers.SignupPageHandler)

	// Password
	app.Get("/password/reset/:verifyId", handlers.ResetPasswordPageHandler)
	app.Post("/password/reset/:verifyId", handlers.ResetPasswordFormHandler)
	app.Get("/password/forget", handlers.ForgetPasswordPageHandler)
	app.Post("/password/forget", handlers.ForgetPasswordFormHandler)

	// Login
	login.Get("/", handlers.LoginPageHandler)
	login.Post("/", handlers.LoginTelarHandler)
	login.Post("/telar", handlers.LoginTelarHandler)
	login.Get("/github", handlers.LoginGithubHandler)
	login.Get("/google", handlers.LoginGoogleHandler)
	app.Get("/oauth2/authorized", handlers.OAuth2Handler)

	// Profile
	app.Put("/profile", authCookieMiddleware, handlers.UpdateProfileHandle)
}
