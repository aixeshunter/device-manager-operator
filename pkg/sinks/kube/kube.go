package kube

import (
	crdClient "hikvision.com/cloud/device-manager/pkg/crd/client/clientset/versioned"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient(kubeConfigFile string) (crdClient.Interface, error) {
	var clientConfig *rest.Config
	var err error
	if len(kubeConfigFile) > 0 {
		clientConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		if err != nil {
			return nil, err
		}
	} else {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err := crdClient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
