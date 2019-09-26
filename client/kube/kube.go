package kube

import (
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultImageTag               = "latest"
	defaultNamespace              = apiv1.NamespaceDefault
	defaultSeleniumSessionIDField = "Selenium"
	kubeApiVersion                = "KUBE_API_VERSION"
	kubeClientName                = "KUBERNETES"
)

var (
	kubeconfig *string
)

type KubeClient struct {
	Type       string
	PodManager v1.PodInterface
}

func CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (*KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podManager := client.CoreV1().Pods(defaultNamespace)
	return &KubeClient{
		Type:       kubeClientName,
		PodManager: podManager,
	}, nil
}
