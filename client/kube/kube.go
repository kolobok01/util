package kube

import (
	"context"
	"io"

	blueclient "github.com/kolobok01/util/client"

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

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podManager := cli.CoreV1().Pods(defaultNamespace)
	return &KubeClient{
		Type:       blueclient.KubeType,
		PodManager: podManager,
	}, nil
}

func (k *KubeClient) GetType() string {
	return k.Type
}

func (k *KubeClient) GetLogs(ctx context.Context, id string) (io.ReadCloser, error) {

}
