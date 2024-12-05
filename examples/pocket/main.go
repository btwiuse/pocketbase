package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/webteleport/relay"
)

func main() {
	if err := Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func Run(args []string) error {
	app := pocketbase.New()

	app.RootCmd.ParseFlags(args)

	// GitHub selfupdate
	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{
		Owner: "btwiuse",
		Repo: "pocketbase",
		ArchiveExecutable: "pocket",
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		app.Logger().Info("starting the relay server", "HOST", apis.HOST)
		store := relay.NewSessionStore()
		mini := relay.NewWSServer(apis.HOST, store)

		withRelay := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isPocketbaseHost := mini.IsRootExternal(r)
				isPocketbaseAPI := strings.HasPrefix(r.URL.Path, "/api/")
				isPocketbase := isPocketbaseHost && isPocketbaseAPI

				if os.Getenv("POCKETBASE_UI") != "" {
					isPocketbaseUI := strings.HasPrefix(r.URL.Path, "/_/")
					isPocketbase = isPocketbaseHost && (isPocketbaseAPI || isPocketbaseUI)
				}

				// route non pocketbase requests to relay
				if !isPocketbase {
					mini.ServeHTTP(w, r)
					return
				}

				next.ServeHTTP(w, r)
			})
		}

		e.Router.Pre(
			apis.ActivityLogger(app),
			echo.WrapMiddleware(withRelay),
		)
		return nil
	})

	return app.Start()
}
