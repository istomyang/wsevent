package publish

import (
	"net/http"
	"wsevent/log"
)

// MiddlewareConfig includes essential info must be used.
type MiddlewareConfig struct {
	Topic string

	// CreateMessage defines a format of a message.
	CreateMessage func(*http.Request) []byte

	// SendCondition judge sending condition.
	SendCondition func(*http.Request) bool

	// Publisher can be assigned to NewKafkaPublisher for default and NewFakeInformer for test, or your customization.
	// Note that you must call Publish.Run and it must be closed when no longer in use.
	Publisher Publish

	// Fail will call when error occurs.
	Fail func(error)
}

func GetMiddlewareFunc(config MiddlewareConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.SendCondition(r) {
			var message = config.CreateMessage(r)
			if err := config.Publisher.Send(message); err != nil {
				config.Fail(err)
				return
			}
			log.Debug("publish-middleware-send: %s", string(message))
		}
	}
}

func InstallStd(config MiddlewareConfig, handler http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		GetMiddlewareFunc(config)(w, r)

		// Next
		handler.ServeHTTP(w, r)
	}
}
