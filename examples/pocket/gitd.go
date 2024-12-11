package main

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/webteleport/ufo/apps/gitd/server"
	"github.com/webteleport/utils"
)

var GitdHook = &hook.Handler[*core.ServeEvent]{
	// static route to serves files from the provided public dir
	// (if publicDir exists and the route path is not already defined)
	Func: func(e *core.ServeEvent) error {
		h := utils.GinLoggerMiddleware(server.Handler())
		handler := apis.WrapStdHandler(h)

		e.Router.GET("/{path...}", handler)
		e.Router.POST("/{path...}", handler)

		return e.Next()
	},

	Priority: 0,
}
