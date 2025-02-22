package postgres

import (
	"fmt"
	"time"

	"github.com/alecthomas/units"

	"github.com/authzed/spicedb/internal/datastore/common"
)

type postgresOptions struct {
	connMaxIdleTime   *time.Duration
	connMaxLifetime   *time.Duration
	healthCheckPeriod *time.Duration
	maxOpenConns      *int
	minOpenConns      *int

	watchBufferLength         uint16
	revisionFuzzingTimedelta  time.Duration
	gcWindow                  time.Duration
	splitAtEstimatedQuerySize units.Base2Bytes

	enablePrometheusStats bool

	logger *tracingLogger
}

const (
	errFuzzingTooLarge = "revision fuzzing timdelta (%s) must be less than GC window (%s)"

	defaultWatchBufferLength = 128
)

// Option provides the facility to configure how clients within the
// Postgres datastore interact with the running Postgres database.
type Option func(*postgresOptions)

func generateConfig(options []Option) (postgresOptions, error) {
	computed := postgresOptions{
		gcWindow:                  24 * time.Hour,
		watchBufferLength:         defaultWatchBufferLength,
		splitAtEstimatedQuerySize: common.DefaultSplitAtEstimatedQuerySize,
	}

	for _, option := range options {
		option(&computed)
	}

	// Run any checks on the config that need to be done
	if computed.revisionFuzzingTimedelta >= computed.gcWindow {
		return computed, fmt.Errorf(
			errFuzzingTooLarge,
			computed.revisionFuzzingTimedelta,
			computed.gcWindow,
		)
	}

	return computed, nil
}

// SplitAtEstimatedQuerySize is the query size at which it is split into two
// (or more) queries.
//
// This value defaults to `common.DefaultSplitAtEstimatedQuerySize`.
func SplitAtEstimatedQuerySize(splitAtEstimatedQuerySize units.Base2Bytes) Option {
	return func(po *postgresOptions) {
		po.splitAtEstimatedQuerySize = splitAtEstimatedQuerySize
	}
}

// ConnMaxIdleTime is the duration after which an idle connection will be
// automatically closed by the health check.
//
// This value defaults to having no maximum.
func ConnMaxIdleTime(idle time.Duration) Option {
	return func(po *postgresOptions) {
		po.connMaxIdleTime = &idle
	}
}

// ConnMaxLifetime is the duration since creation after which a connection will
// be automatically closed.
//
// This value defaults to having no maximum.
func ConnMaxLifetime(lifetime time.Duration) Option {
	return func(po *postgresOptions) {
		po.connMaxLifetime = &lifetime
	}
}

// HealthCheckPeriod is the interval by which idle Postgres client connections
// are health checked in order to keep them alive in a connection pool.
func HealthCheckPeriod(period time.Duration) Option {
	return func(po *postgresOptions) {
		po.healthCheckPeriod = &period
	}
}

// MaxOpenConns is the maximum size of the connection pool.
//
// This value defaults to having no maximum.
func MaxOpenConns(conns int) Option {
	return func(po *postgresOptions) {
		po.maxOpenConns = &conns
	}
}

// MinOpenConns is the minimum size of the connection pool.
// The health check will increase the number of connections to this amount if
// it had dropped below.
//
// This value defaults to zero.
func MinOpenConns(conns int) Option {
	return func(po *postgresOptions) {
		po.minOpenConns = &conns
	}
}

// WatchBufferLength is the number of entries that can be stored in the watch
// buffer while awaiting read by the client.
//
// This value defaults to 128.
func WatchBufferLength(watchBufferLength uint16) Option {
	return func(po *postgresOptions) {
		po.watchBufferLength = watchBufferLength
	}
}

// RevisionFuzzingTimedelta is the time bucket size to which advertised
// revisions will be rounded.
//
// This value defaults to 5 seconds.
func RevisionFuzzingTimedelta(delta time.Duration) Option {
	return func(po *postgresOptions) {
		po.revisionFuzzingTimedelta = delta
	}
}

// GCWindow is the maximum age of a passed revision that will be considered
// valid.
//
// This value defaults to 24 hours.
func GCWindow(window time.Duration) Option {
	return func(po *postgresOptions) {
		po.gcWindow = window
	}
}

// EnablePrometheusStats enables Prometheus metrics provided by the Postgres
// clients being used by the datastore.
func EnablePrometheusStats() Option {
	return func(po *postgresOptions) {
		po.enablePrometheusStats = true
	}
}

// EnableTracing enables trace-level logging for the Postgres clients being
// used by the datastore.
func EnableTracing() Option {
	return func(po *postgresOptions) {
		po.logger = &tracingLogger{}
	}
}
