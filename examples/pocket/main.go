package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/webteleport/relay"
	"github.com/webteleport/utils"
)

func main() {
	app := pocketbase.New()

	app.RootCmd.ParseFlags(os.Args[1:])

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		host := utils.EnvHost("localhost")
		app.Logger().Info("starting the relay server", "HOST", host)
		store := relay.NewSessionStore()
		mini := relay.NewWSServer(host, store)

		withRelay := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isPocketbaseHost := mini.IsRoot(r)
				isPocketbaseAPI := strings.HasPrefix(r.URL.Path, "/api/")
				isPocketbase := isPocketbaseHost && isPocketbaseAPI
				// route request to the relay server
				if !isPocketbase {
					mini.ServeHTTP(w, r)
					return
				}
				// route request to the pocketbase api
				next.ServeHTTP(w, r)
			})
		}

		e.Router.Pre(
			apis.ActivityLogger(app),
			echo.WrapMiddleware(withRelay),
		)
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
