package client

type Client interface {
	GetType() string
	CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (Client, error)
}

const (
	DockerType = "DOCKER"
	KubeType   = "KUBERNETES"
)
