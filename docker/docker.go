package docker

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api"
	"github.com/docker/docker/client"

	"context"
	"strconv"
)

const dockerApiVersion = "DOCKER_API_VERSION"

func CreateCompatibleDockerClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (*client.Client, error) {
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
				docker, err := client.NewClientWithOpts(client.FromEnv)
				if err != nil {
					return nil, err
				}
				if isDockerAPIVersionCorrect(docker) {
					onVersionDetermined(apiVersion)
					return docker, nil
				}
				docker.Close()
			}
		}
		onUsingDefaultVersion(api.DefaultVersion)
	}
	return client.NewClientWithOpts(client.FromEnv)
}

func isDockerAPIVersionCorrect(docker *client.Client) bool {
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
