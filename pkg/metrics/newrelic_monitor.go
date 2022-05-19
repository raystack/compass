package metrics

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type NewRelicConfig struct {
	Enabled    bool   `mapstructure:"enabled" default:"false"`
	AppName    string `mapstructure:"appname" default:"compass"`
	LicenseKey string `mapstructure:"licensekey" default:""`
}

type NewRelicMonitor struct {
	app *newrelic.Application
}

func NewNewRelicMonitor(app *newrelic.Application) *NewRelicMonitor {
	return &NewRelicMonitor{
		app: app,
	}
}

func (mon *NewRelicMonitor) Application() *newrelic.Application {
	if mon != nil {
		return mon.app
	}
	return nil
}

func (mon *NewRelicMonitor) MonitorRouter(router *mux.Router) {
	router.Use(nrgorilla.Middleware(mon.app))

	// below handlers still have to be manually wrapped by newrelic core library
	_, router.NotFoundHandler = newrelic.WrapHandle(mon.app, "NotFoundHandler", router.NotFoundHandler)
	_, router.MethodNotAllowedHandler = newrelic.WrapHandle(mon.app, "MethodNotAllowedHandler", router.MethodNotAllowedHandler)
}

func (mon *NewRelicMonitor) StartTransaction(ctx context.Context, operation string) (context.Context, func()) {
	txn := mon.app.StartTransaction(operation)
	ctx = newrelic.NewContext(ctx, txn)

	return ctx, func() {
		txn.End()
	}
}
