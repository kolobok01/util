package client

type Client interface {
	GetType() string
}

const (
	DockerType = "DOCKER"
	KubeType   = "KUBERNETES"
)
