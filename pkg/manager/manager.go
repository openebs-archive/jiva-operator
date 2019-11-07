package manager

import "time"

var (
	// SyncPeriod is the time after which reconcilation
	// will be triggered periodically
	SyncPeriod = 5 * time.Second
	// RetryPeriod is the time after which reconcilation
	// will be retried in case it was failed in previous attempt
	RetryPeriod = 2 * time.Second
)
