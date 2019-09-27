package client

import (
	"context"
	"io"
)

type Client interface {
	GetType() string
	GetLogs(ctx context.Context, container string) (io.ReadCloser, error)
}

const (
	DockerType = "DOCKER"
	KubeType   = "KUBERNETES"
)
