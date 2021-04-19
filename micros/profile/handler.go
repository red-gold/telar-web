package function

import (
	"context"
	"fmt"
	"net/http"

	coreServer "github.com/red-gold/telar-core/server"
	micros "github.com/red-gold/telar-web/micros"
	"github.com/red-gold/telar-web/micros/profile/handlers"
)

func init() {

	micros.InitConfig()
}

// Cache state
var server *coreServer.ServerRouter
var db interface{}

// Handler function
func Handle(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	// Start
	if db == nil {
		var startErr error

		db, startErr = micros.Start(ctx)

		if startErr != nil {
			fmt.Printf("Error startup: %s", startErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(startErr.Error()))
		}
	}

	// Server Routing
	if server == nil {
		fmt.Println("Server is nil")
		server = coreServer.NewServerRouter()
		server.GET("/my", handlers.ReadMyProfileHandle(db), coreServer.RouteProtectionCookie)
		server.GETWR("/", handlers.QueryUserProfileHandle(db), coreServer.RouteProtectionCookie)
		server.GET("/id/:userId", handlers.ReadProfileHandle(db), coreServer.RouteProtectionCookie)
		server.POST("/index", handlers.InitProfileIndexHandle(db), coreServer.RouteProtectionHMAC)
		server.PUT("/last-seen", handlers.UpdateLastSeen(db), coreServer.RouteProtectionHMAC)

		// Invoke between functions and protected by HMAC
		server.PUTWR("/", handlers.UpdateProfileHandle(db), coreServer.RouteProtectionHMAC)
		server.GET("/dto/id/:userId", handlers.ReadDtoProfileHandle(db), coreServer.RouteProtectionHMAC)
		server.POST("/dto", handlers.CreateDtoProfileHandle(db), coreServer.RouteProtectionHMAC)
		server.POST("/dispatch", handlers.DispatchProfilesHandle(db), coreServer.RouteProtectionHMAC)
		server.POST("/dto/ids", handlers.GetProfileByIds(db), coreServer.RouteProtectionHMAC)
	}
	server.ServeHTTP(w, r)
}
