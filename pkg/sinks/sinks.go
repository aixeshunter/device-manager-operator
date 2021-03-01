// Package sinks is an interface that store hardware information to different storage.
package sinks

import (
	crdClient "hikvision.com/cloud/device-manager/pkg/crd/client/clientset/versioned"
	"hikvision.com/cloud/device-manager/pkg/sinks/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
