package docker

import (
	"fmt"
	"io"
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
						dockerClientName,
						docker,
					}, nil
				}
				docker.Close()
			}
		}
		onUsingDefaultVersion(api.DefaultVersion)
	}
	cl, err := client.NewEnvClient()
	return &DockerClient{
		blueclient.DockerType,
		cl,
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
	return d.Type
}

func (d *DockerClient) GetLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{}
	return d.Client.ContainerLogs(ctx, id, options)
}

func (d *DockerClient) SetDebug(debug bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.debug = debug
}
