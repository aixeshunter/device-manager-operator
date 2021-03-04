// Package sinks is an interface that store hardware information to different storage.
package sinks

import (
	"context"
	crdClient "hikvision.com/cloud/device-manager/pkg/crd/client/clientset/versioned"
	"hikvision.com/cloud/device-manager/pkg/sinks/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"time"
)

type SinkProvider struct {
	kubeClient  kubernetes.Interface
	client      crdClient.Interface
	maxRetry    int
	retryPeriod time.Duration
	nodeName    string
	chroot      string
}

func NewSinkProvider(kubeConfig string, maxRetry int, retryPeriod time.Duration, nodeName string, chroot string) (*SinkProvider, error) {
	client, err := kube.NewClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &SinkProvider{
		kubeClient:  kubeClient,
		client:      client,
		maxRetry:    maxRetry,
		retryPeriod: retryPeriod,
		nodeName:    nodeName,
		chroot:      chroot,
	}, nil
}

func (p *SinkProvider) DiskHandler(ctx context.Context) error {
	klog.V(5).Infof("Handler the disks starting...")
	var err error
	retry := 0
	for retry < p.maxRetry {
		select {
		case <-ctx.Done():
			klog.V(1).Infof("Exit sink handler.")
			return ctx.Err()
		default:
			if err = kube.HandleExtendDevice(ctx, p.client, p.nodeName, p.chroot); err != nil {
				klog.Warning("Failed to handler the disks.")
			} else {
				return nil
			}
			retry++
			if retry != p.maxRetry {
				time.Sleep(p.retryPeriod)
			}
		}
	}

	if retry == p.maxRetry {
		klog.Errorf("Failed to handler disks after %d retries.", p.maxRetry)
		return err
	}
	klog.V(2).Infoln("handler devices successfully.")

	return nil
}
