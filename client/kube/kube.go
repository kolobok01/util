package kube

import (
	"bytes"
	"context"
	"fmt"
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
		log.Printf("DEBUG: Starting GetLogs for ID, context: %s, %+v", id, ctx)
	}

	req := k.PodManager.GetLogs(id, nil)
	res := req.Do()
	logs, err := res.Raw()
	if err != nil {
		if k.debug {
			log.Printf("DEBUG: error in GetLogs for ID %s: %s", id, err.Error())
		}
		return nil, err
	}

	r := ioutil.NopCloser(bytes.NewReader(logs))
	if k.debug {
		log.Printf("DEBUG: success in GetLogs for ID %s: %b", id, logs)
	}
	return r, nil
}

func (k *KubeClient) SetDebug(debug bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.debug = debug
}

func (k *KubeClient) LaunchPod(name string, podSpec *apiv1.PodSpec) (*apiv1.Pod, error) {
	if k.debug {
		log.Printf("DEBUG: LaunchPod: PodSpec: %+v", podSpec)
	}

	return k.PodManager.Create(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: *podSpec,
	})
}

func BuildSessionPod(requestId, image string) *apiv1.Pod {
	return &apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: apiv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    defaultNamespace,
			GenerateName: "blueio-",
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyNever,
			Containers: []apiv1.Container{
				{
					Name:  fmt.Sprintf("%s-%s", image, requestId),
					Image: fmt.Sprintf("expertio/vnc:%s", image),
				},
			},
		},
	}
}

func (k *KubeClient) CreateSessionPod(requestId, image string) (*apiv1.Pod, error) {
	if k.debug {
		log.Printf("DEBUG: CreateSessionPod: requestId, image: %s, %s", requestId, image)
	}

	pod := BuildSessionPod(requestId, image)
	return k.CreatePod(pod)
}

func (k *KubeClient) CreatePod(pod *apiv1.Pod) (*apiv1.Pod, error) {
	if k.debug {
		log.Printf("DEBUG: CreatePod: podspec: %+v", pod)
	}

	return k.PodManager.Create(pod)
}

func (k *KubeClient) DeletePodByName(name string) error {
	if k.debug {
		log.Printf("DEBUG: DeletePodByName: name: %s", name)
	}

	return k.PodManager.Delete(name, &metav1.DeleteOptions{
		// TODO: get proper options
	})
}

func (k *KubeClient) GetPodByName(name string) (*apiv1.Pod, error) {
	if k.debug {
		log.Printf("DEBUG: GetPodByName: name: %s", name)
	}

	return k.PodManager.Get(name, metav1.GetOptions{})
}

func (k *KubeClient) GetPodBySessionID(name string) (*apiv1.Pod, error) {
	if k.debug {
		log.Printf("DEBUG: GetPodBySessionID: name: %s", name)
	}

	return k.PodManager.Get(name, metav1.GetOptions{})
}

func (k *KubeClient) AddSessionID(name, sessionID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	pod, err := k.GetPodByName(name)
	if err != nil {
		return err
	}
	current := pod.Labels[defaultSeleniumSessionIDField]
	if current != "" {
		return fmt.Errorf("%s is already set on pod %s", defaultSeleniumSessionIDField, name)
	}

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[defaultSeleniumSessionIDField] = sessionID

	_, err = k.PodManager.Update(pod)
	return err

}
