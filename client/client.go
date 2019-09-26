package client

import (
	"github.com/docker/docker/client"
)

type Client interface {
	CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (*client.Client, error)
}
