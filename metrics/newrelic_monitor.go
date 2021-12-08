package metrics

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type NewrelicMonitor struct {
	app *newrelic.Application
}

func NewNewrelicMonitor(app *newrelic.Application) *NewrelicMonitor {
	return &NewrelicMonitor{
		app: app,
	}
}

func (mon *NewrelicMonitor) MonitorRouter(router *mux.Router) {
	router.Use(nrgorilla.Middleware(mon.app))

	// below handlers still have to be manually wrapped by newrelic core library
	_, router.NotFoundHandler = newrelic.WrapHandle(mon.app, "NotFoundHandler", router.NotFoundHandler)
	_, router.MethodNotAllowedHandler = newrelic.WrapHandle(mon.app, "MethodNotAllowedHandler", router.MethodNotAllowedHandler)
}

func (mon *NewrelicMonitor) StartTransaction(ctx context.Context, operation string) (context.Context, func()) {
	txn := mon.app.StartTransaction(operation)
	ctx = newrelic.NewContext(ctx, txn)

	return ctx, func() {
		txn.End()
	}
}
