package function

import (
	"net/http"

	coreServer "github.com/red-gold/telar-core/server"
	appConfig "github.com/red-gold/telar-web/micros/storage/config"
	"github.com/red-gold/telar-web/micros/storage/handlers"
)

func init() {
	appConfig.InitConfig()
}

// Cache state
var server *coreServer.ServerRouter

// Handler function
func Handle(w http.ResponseWriter, r *http.Request) {

	// Server Routing
	if server == nil {
		server = coreServer.NewServerRouter()
		server.POSTFILE("/:uid/:dir", handlers.UploadeHandle(), coreServer.RouteProtectionCookie)
		// server.PUTWR("/", handlers.GetFileHandle(db), coreServer.RouteProtectionCookie)
		// server.DELETEWR("/file/:name", handlers.GetFileHandle(db), coreServer.RouteProtectionCookie)
		// server.DELETEWR("/dir/:dir", handlers.GetFileHandle(db), coreServer.RouteProtectionCookie)
		server.GETWR("/:uid/:dir/:name", handlers.GetFileHandle(), coreServer.RouteProtectionCookie)
	}
	server.ServeHTTP(w, r)
}
