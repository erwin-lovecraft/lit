package monitoring

import (
	"fmt"
	"maps"
	"reflect"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
)

// DefaultFlushWait represents the default wait time for flushing
const DefaultFlushWait = 10 * time.Second // 10 sec (this flush) + 10 sec for server shutdown grace = 20 sec which is reasonable considering k8s grace is 30sec

// Monitor represents instance for logging and capture error
type Monitor struct {
	sentryClient *sentry.Client
	logger       *zap.Logger
	// Currently unable to retrieve logTags saved in uber zap logger due to its design to be quick.
	// Hence, keeping a local copy of logTags for other purpose such as sentry error reporting
	logTags map[string]string
}

func (m *Monitor) WithTag(key string, value string) *Monitor {
	if m == nil {
		return nil
	}

	clonedTags := maps.Clone(m.logTags)
	clonedTags[key] = value
	return &Monitor{
		sentryClient: m.sentryClient,
		logger:       m.logger,
		logTags:      clonedTags,
	}
}

// With creates a new child Monitor and adds new logTags to it. Parent Monitor remains unchanged.
func (m *Monitor) With(tags map[string]string) *Monitor {
	if m == nil {
		return nil
	}

	// Add new log field
	clonedTags := maps.Clone(m.logTags)
	for k, v := range tags {
		clonedTags[k] = v
	}

	return &Monitor{
		sentryClient: m.sentryClient,
		logger:       m.logger,
		logTags:      clonedTags,
	}
}

func (m *Monitor) getLogFields() []zap.Field {
	if m == nil {
		return nil
	}

	logFields := make([]zap.Field, 0, len(m.logTags))
	for k, v := range m.logTags {
		logFields = append(logFields, zap.String(k, v))
	}
	return logFields
}

// Infof logs the message using info level
func (m *Monitor) Infof(format string, args ...interface{}) {
	if m == nil {
		return
	}

	m.logger.Info(fmt.Sprintf(format, args...), m.getLogFields()...)
}

// Errorf logs the message using error level and reports the error to sentry
func (m *Monitor) Errorf(err error, msg string, args ...interface{}) {
	if m == nil {
		return
	}

	logFields := m.getLogFields()
	// Capture error.key, error.message and error.stack to log
	logFields = append(logFields,
		zap.String("error.kind", reflect.TypeOf(err).String()),
		zap.String("error.message", err.Error()),
	)

	if v, ok := err.(stackTracer); ok {
		stack := fmt.Sprintf("%+v", v.StackTrace())
		if len(stack) > 0 && stack[0] == '\n' {
			stack = stack[1:]
		}
		logFields = append(logFields, zap.String("error.stack", stack))
	}

	if msg != "" {
		m.logger.Error(fmt.Sprintf(msg+". Err: %v", append(args, err)...), logFields...)
	} else {
		m.logger.Error(fmt.Sprintf("Err: %v", err), logFields...)
	}

	m.ReportError(err, m.logTags)
}

func (m *Monitor) ReportError(err error, tags map[string]string) {
	if m == nil || m.sentryClient == nil {
		return
	}
	scope := sentry.NewScope()
	scope.SetTags(tags)
	m.sentryClient.CaptureException(err, nil, scope)
}

// Flush will flush all the monitor data left in the queue to the monitoring service
func (m *Monitor) Flush(maxWait time.Duration) {
	if m == nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)

	// Zap
	go func() {
		defer wg.Done()
		_ = flushZap(m.logger, DefaultFlushWait)
	}()

	// Sentry
	go func() {
		defer wg.Done()
		if m.sentryClient != nil {
			m.sentryClient.Flush(maxWait)
		}
	}()

	wg.Wait() // In the worst case, the pod may be killed before clearing this statement, but it's not critical.
}

// This is to extract the stack trace from pkgerrors
type stackTracer interface {
	StackTrace() pkgerrors.StackTrace
}
