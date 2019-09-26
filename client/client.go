package client

type Client interface {
	Type() string
	CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (Client, error)
}

const (
	DockerType = "DOCKER"
	KubeType   = "KUBERNETES"
)
