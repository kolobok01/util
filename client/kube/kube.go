package kube

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	blueclient "github.com/kolobok01/util/client"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	kubeconfig string
)

type KubeClient struct {
	Type       string
	PodManager v1.PodInterface
	debug      bool
	mu         sync.Mutex
}

func CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (*KubeClient, error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dcli := cli.DiscoveryClient
	v, err := dcli.ServerVersion()
	if err != nil {
		return nil, err
	} else {
		onVersionDetermined(v.String())
	}

	podManager := cli.CoreV1().Pods(defaultNamespace)
	return &KubeClient{
		Type:       blueclient.KubeType,
		PodManager: podManager,
	}, nil
}

func (k *KubeClient) GetType() string {
	if k.debug {
		log.Printf("DEBUG: type is %s", k.Type)
	}
	return k.Type
}

func (k *KubeClient) GetLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	if k.debug {
		log.Printf("DEBUG: Starting GetLogs for ID: %s", id)
		log.Printf("DEBUG: context %+v", ctx)
	}
	req := k.PodManager.GetLogs(id, nil)
	res := req.Do()
	logs, err := res.Raw()
	if err != nil {
		if k.debug {
			log.Printf("DEBUG: error in GetLogs for ID: %s", id)
			log.Printf("DEBUG: error %s", err.Error())
		}
		return nil, err
	}
	r := ioutil.NopCloser(bytes.NewReader(logs))
	if k.debug {
		log.Printf("DEBUG: success in GetLogs for ID: %s", id)
		log.Printf("DEBUG: logs: %b", logs)
	}
	return r, nil
}

func (k *KubeClient) SetDebug(debug bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.debug = debug
}

func (k *KubeClient) LaunchPod(name string, podSpec *apiv1.PodSpec) error {
	if k.debug {
		log.Printf("DEBUG: LaunchPod: PodSpec: %+v", podSpec)
	}

	_, err := k.PodManager.Create(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: *podSpec,
	})
	if k.debug {
		log.Printf("DEBUG: LaunchPod: err: %+v", err)
	}
	return err
}
