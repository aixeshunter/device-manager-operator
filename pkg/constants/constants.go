package constants

import "time"

const (
	DefaultPeriod      = 30 * time.Second
	DefaultMaxRetry    = 1
	DefaultRetryPeriod = 10 * time.Second

	NodeNameEnv = "DEVICE_NODE_NAME"
)
