package platform

import (
	"time"

	"github.com/getsentry/sentry-go"
)

// InitSentry initializes the Sentry SDK. If dsn is empty, Sentry is disabled
// and a no-op cleanup function is returned.
func InitSentry(dsn, environment, release string) (cleanup func(), err error) {
	noop := func() {}
	if dsn == "" {
		return noop, nil
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          release,
		TracesSampleRate: 0.1,
	})
	if err != nil {
		return noop, err
	}

	return func() { sentry.Flush(2 * time.Second) }, nil
}
