package constants

import "time"

const (
	DefaultPeriod      = 10 * time.Second
	DefaultMaxRetry    = 1
	DefaultRetryPeriod = 5 * time.Second

	NodeNameEnv = "DEVICE_NODE_NAME"

	TimeLayout = "2006-01-02 15:04:05"
)
