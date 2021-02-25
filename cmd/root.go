package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"hikvision.com/cloud/device-manager/cmd/app"

	"k8s.io/component-base/logs"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	logs.InitLogs()
	defer logs.FlushLogs()

	c := app.NewDeviceManagerCommand()
	if err := c.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
