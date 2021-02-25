package constants

import "time"

const (
	DefaultPeriod      = 10 * time.Second
	DefaultMaxRetry    = 3
	DefaultRetryPeriod = 5 * time.Second

	NodeNameEnv = "DEVICE_NODE_NAME"
)
