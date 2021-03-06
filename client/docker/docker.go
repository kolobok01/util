package docker

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	blueclient "github.com/kolobok01/util/client"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"context"
	"strconv"
)

const (
	dockerApiVersion = "DOCKER_API_VERSION"
	dockerClientName = "DOCKER"
)

type DockerClient struct {
	Type   string
	Client *client.Client
	debug  bool
	mu     sync.Mutex
}

func CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (*DockerClient, error) {
	dockerApiVersionEnv := os.Getenv(dockerApiVersion)
	if dockerApiVersionEnv != "" {
		onVersionSpecified(dockerApiVersionEnv)
	} else {
		maxMajorVersion, maxMinorVersion := parseVersion(api.DefaultVersion)
		minMajorVersion, minMinorVersion := parseVersion(api.MinVersion)
		for majorVersion := maxMajorVersion; majorVersion >= minMajorVersion; majorVersion-- {
			for minorVersion := maxMinorVersion; minorVersion >= minMinorVersion; minorVersion-- {
				apiVersion := fmt.Sprintf("%d.%d", majorVersion, minorVersion)
				os.Setenv(dockerApiVersion, apiVersion)
				docker, err := client.NewEnvClient()
				if err != nil {
					return nil, err
				}
				if isAPIVersionCorrect(docker) {
					onVersionDetermined(apiVersion)
					return &DockerClient{
						Type:   blueclient.DockerType,
						Client: docker,
					}, nil
				}
				docker.Close()
			}
		}
		onUsingDefaultVersion(api.DefaultVersion)
	}
	cl, err := client.NewEnvClient()
	return &DockerClient{
		Type:   blueclient.DockerType,
		Client: cl,
	}, err
}

func isAPIVersionCorrect(docker *client.Client) bool {
	ctx := context.Background()
	apiInfo, err := docker.ServerVersion(ctx)
	if err != nil {
		return false
	}
	return apiInfo.APIVersion == docker.ClientVersion()
}

func parseVersion(ver string) (int, int) {
	const point = "."
	pieces := strings.Split(ver, point)
	major, err := strconv.Atoi(pieces[0])
	if err != nil {
		return 0, 0
	}
	minor, err := strconv.Atoi(pieces[1])
	if err != nil {
		return 0, 0
	}
	return major, minor
}

func (d *DockerClient) GetType() string {
	if d.debug {
		log.Printf("[%d] [DOCKER_GET_TYPE] [TYPE: %s]", 0, d.Type)
	}
	return d.Type
}

func (d *DockerClient) GetLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	if d.debug {
		log.Printf("[%d] [DOCKER_GET_LOGS] [ID: %s]", 0, id)
	}
	options := types.ContainerLogsOptions{}
	return d.Client.ContainerLogs(ctx, id, options)
}

func (d *DockerClient) SetDebug(debug bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.debug = debug
	if d.debug {
		log.Printf("[%d] [DOCKER_DEBUG] [DEBUG: %s]", 0, d.debug)
	}
}
