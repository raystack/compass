package metrics

import (
	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type NewrelicMonitor struct {
	app       *newrelic.Application
	prefix    string
	separator string
}

func NewNewrelicMonitor(app *newrelic.Application) *NewrelicMonitor {
	return &NewrelicMonitor{
		app: app,
	}
}

func (mm *NewrelicMonitor) MonitorRouter(router *mux.Router) {
	router.Use(nrgorilla.Middleware(mm.app))

	// below handlers still have to be manually wrapped by newrelic core library
	_, router.NotFoundHandler = newrelic.WrapHandle(mm.app, "NotFoundHandler", router.NotFoundHandler)
	_, router.MethodNotAllowedHandler = newrelic.WrapHandle(mm.app, "MethodNotAllowedHandler", router.MethodNotAllowedHandler)
}
