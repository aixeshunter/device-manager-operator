package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"hikvision.com/cloud/device-manager/pkg/constants"
	"hikvision.com/cloud/device-manager/pkg/sinks"
	"hikvision.com/cloud/device-manager/pkg/version"

	"k8s.io/apimachinery/pkg/util/wait"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/klog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// options command line flags.
type options struct {
	Period         time.Duration
	MaxRetry       int
	RetryPeriod    time.Duration
	KubeConfigFile string
	NodeName       string
}

func newOptions() *options {
	return &options{
		Period:      constants.DefaultPeriod,
		MaxRetry:    constants.DefaultMaxRetry,
		RetryPeriod: constants.DefaultRetryPeriod,
	}
}

func (o *options) Flags() (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet("nhd")
	fs.DurationVar(&o.Period, "period", o.Period, "The duration that should wait between attempting acquisition hardware information.")
	fs.IntVar(&o.MaxRetry, "max-retry", o.MaxRetry, "The max count that will retry if store data to Kubernetes failed.")
	fs.DurationVar(&o.RetryPeriod, "retry-period", o.RetryPeriod, "The period that should wait between retry store data to Kubernetes")
	fs.StringVar(&o.KubeConfigFile, "kubeconfig", o.KubeConfigFile, "The KubeConfig file to use when talking to the cluster.")
	return
}

func NewDeviceManagerCommand() *cobra.Command {
	opts := newOptions()
	var chroot string

	cmd := &cobra.Command{
		Use:  "device-manager",
		Long: `The device-manager is a tool to manage devices of node, like disk, usb etc.`,
		Run: func(cmd *cobra.Command, args []string) {
			klog.Infof("Starting vice-manager %v.", version.Get())

			nodeName := os.Getenv(constants.NodeNameEnv)
			if nodeName == "" {
				_, _ = fmt.Fprintf(os.Stderr, "The env %s is empty.\n", constants.NodeNameEnv)
				os.Exit(1)
			}
			opts.NodeName = nodeName

			if err := runCommand(opts, chroot); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	fs.StringVar(&chroot, "chroot", "/", "The the root host filesystem bind-mounted to the mount point.")

	return cmd
}

// run executes the logic. It only returns on error or context is done.
func run(ctx context.Context, chroot string, period time.Duration, provider *sinks.SinkProvider) error {

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		klog.Infoln("The work queue starting...")
		//data, err := discovery.Collect(chroot)
		//if err != nil {
		//	klog.Errorf("Failed to collect hardware information, reason: %v", err)
		//	return
		//}
		//
		//if err := provider.Store(ctx, data); err != nil {
		//	klog.Errorf("Failed to store hardware information, reason: %v", err)
		//	return
		//}

		klog.Infoln("The work queue done. Waiting the next queue.")
	}, period)

	return nil
}

func runCommand(opts *options, chroot string) error {
	klog.V(1).Infof("The period that manager devices is %v.", opts.Period)
	klog.V(1).Infof("Retry count is %d, period is %v if store fail.", opts.MaxRetry, opts.RetryPeriod)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go listenToSystemSignal(cancel)

	sp, err := sinks.NewSinkProvider(opts.KubeConfigFile, opts.MaxRetry, opts.RetryPeriod, opts.NodeName, chroot)
	if err != nil {
		klog.Errorf("Failed to initialize sink provider, reason: %v", err)
		return err
	}

	return run(ctx, chroot, opts.Period, sp)
}

// listenToSystemSignal listen system signal and exit.
func listenToSystemSignal(cancel context.CancelFunc) {
	klog.V(3).Info("Listen to system signal.")

	signalChan := make(chan os.Signal, 1)
	ignoreChan := make(chan os.Signal, 1)

	signal.Notify(ignoreChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		klog.Infof("Shutdown by system signal: %s", sig)
		cancel()
		return
	}
}
