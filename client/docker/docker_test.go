package docker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/aandryashin/matchers"
	"github.com/docker/docker/api"
	"github.com/kolobok01/util"
)

var (
	mockDockerServer *httptest.Server
	apiVersion       = "1.29"
)

func init() {
	mockDockerServer = httptest.NewServer(mux())
	os.Setenv("DOCKER_HOST", "tcp://"+util.HostPort(mockDockerServer.URL))
}

func mux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/v%s/version", apiVersion), http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			output := fmt.Sprintf(`
				{
				
					"Version": "17.04.0",
					"Os": "linux",
					"KernelVersion": "3.19.0-23-generic",
					"GoVersion": "go1.7.5",
					"GitCommit": "deadbee",
					"Arch": "amd64",
					"ApiVersion": "%s",
					"MinAPIVersion": "1.12",
					"BuildTime": "2016-06-14T07:09:13.444803460+00:00",
					"Experimental": true
				
				}
			`, apiVersion)
			w.Write([]byte(output))
		},
	))
	return mux
}

func TestCreateCompatibleDockerClient(t *testing.T) {
	testCreateCompatibleDockerClient(t, "1.27")
}

func TestCreateDockerClientVersionSpecified(t *testing.T) {
	os.Setenv("DOCKER_API_VERSION", "1.27")
	defer os.Unsetenv("DOCKER_API_VERSION")
	testCreateCompatibleDockerClient(t, "1.27")
}

func TestCreateDockerClientDefaultVersion(t *testing.T) {
	major, minor := parseVersion(api.DefaultVersion)
	apiVersion = fmt.Sprintf("%d.%d", major, minor+1)
	defer func() {
		apiVersion = "1.29"
	}()
	testCreateCompatibleDockerClient(t, api.DefaultVersion)
}

func testCreateCompatibleDockerClient(t *testing.T, determinedVersion string) {
	var version string
	fn := func(v string) {
		version = v
	}
	cli, err := CreateCompatibleClient(fn, fn, fn)
	AssertThat(t, err, Is{nil})
	AssertThat(t, cli, Not{nil})
	AssertThat(t, version, EqualTo{determinedVersion})
}

func TestParseCorrectVersion(t *testing.T) {
	major, minor := parseVersion("42.33")
	AssertThat(t, major, EqualTo{42})
	AssertThat(t, minor, EqualTo{33})
}

func TestParseIncorrectMajorVersion(t *testing.T) {
	major, minor := parseVersion("a.22")
	AssertThat(t, major, EqualTo{0})
	AssertThat(t, minor, EqualTo{0})
}

func TestParseIncorrectMinorVersion(t *testing.T) {
	major, minor := parseVersion("1.b")
	AssertThat(t, major, EqualTo{0})
	AssertThat(t, minor, EqualTo{0})
}
