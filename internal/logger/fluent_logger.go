package logger

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/go-kit/kit/log"
)

type fluentLogger struct {
	// log.Logger
	Driver *fluent.Fluent
}

// NewFluentLogger to send to fluentd
func NewFluentLogger() log.Logger {
	driver, _ := fluent.New(fluent.Config{Async: true})
	return &fluentLogger{Driver: driver}
}

func (l *fluentLogger) Log(keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	var tag string
	var message = make(map[string]interface{})

	for i := 0; i < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]

		if k == "tag" {
			tag = v.(string)
		} else {
			message[k.(string)] = v
		}
	}

	return l.Driver.Post(tag, message)
}

func (l *fluentLogger) Close() {
	l.Driver.Close()
}
