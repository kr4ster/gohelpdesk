package route

import (
	"net/http"

	"github.com/kr4ster/gohelpdesk/controller"
	"github.com/kr4ster/gohelpdesk/route/middleware/acl"
	hr "github.com/kr4ster/gohelpdesk/route/middleware/httproutewrapper"
	"github.com/kr4ster/gohelpdesk/route/middleware/logrequest"
	"github.com/kr4ster/gohelpdesk/route/middleware/pprofhandler"
	"github.com/kr4ster/gohelpdesk/shared/session"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/kr4ster/csrfbanana"
)

func Load() http.Handler {
	return middleware(routes())
}

func LoadHTTPS() http.Handler {
	return middleware(routes())
}

func LoadHTTP() http.Handler {
	return middleware(routes())
}

func redirectToHTTPS(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "https://"+req.Host, http.StatusMovedPermanently)
}

func routes() *httprouter.Router {
	r := httprouter.New()

	r.NotFound = alice.
		New().
		ThenFunc(controller.Error404)

	r.GET("/static/*filepath", hr.Handler(alice.
		New().
		ThenFunc(controller.Static)))

	// Home page
	r.GET("/", hr.Handler(alice.
		New().
		ThenFunc(controller.Index)))

	// Login
	/*
		r.GET("/login", hr.Handler(alice.
			New(acl.DisallowAuth).
			ThenFunc(controller.LoginGET)))
		r.POST("/login", hr.Handler(alice.
			New(acl.DisallowAuth).
			ThenFunc(controller.LoginPOST)))
		r.GET("/logout", hr.Handler(alice.
			New().
			ThenFunc(controller.Logout)))
	*/

	// Register
	/*
		r.GET("/register", hr.Handler(alice.
			New(acl.DisallowAuth).
			ThenFunc(controller.RegisterGET)))
		r.POST("/register", hr.Handler(alice.
			New(acl.DisallowAuth).
			ThenFunc(controller.RegisterPOST)))
	*/

	// About
	r.GET("/about", hr.Handler(alice.
		New().
		ThenFunc(controller.AboutGET)))

	// Enable Pprof
	r.GET("/debug/pprof/*pprof", hr.Handler(alice.
		New(acl.DisallowAnon).
		ThenFunc(pprofhandler.Handler)))

	return r
}

func middleware(h http.Handler) http.Handler {
	// Prevent double submits
	cs := csrfbanana.New(h, session.Store, session.Name)
	cs.FailureHandler(http.HandlerFunc(controller.InvalidToken))
	cs.ClearAfterUsage(true)
	cs.ExcludeRegexPaths([]string{"/static(.*)"})
	csrfbanana.TokenLength = 32
	csrfbanana.TokenName = "token"
	csrfbanana.SingleToken = false
	h = cs

	// Log every request
	h = logrequest.Handler(h)

	// Clear handler for Gorilla Context
	h = context.ClearHandler(h)

	return h
}
