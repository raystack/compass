package testutil

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	elasticSearchCmdLine = []string{
		"docker", "run",
		"-d", "-P", "--rm",
		"-e", "discovery.type=single-node",
		"-e", "path.data=/opt/elasticsearch/volatile/data",
		"-e", "path.logs=/opt/elasticsearch/volatile/logs",
		"-e", "bootstrap.memory_lock=true",
		"-e", "ES_JAVA_OPTS=-Xms128m -Xmx128m -server",
		"-e", "ES_HEAP_SIZE=128m",
		"-e", "MAX_LOCKED_MEMORY=100000",
		"-e", "xpack.security.enabled=false",
		"--memory-swappiness=0",
		"--memory-swap=0",
		"--tmpfs", "/opt/elasticsearch/volatile/data:rw",
		"--tmpfs", "/opt/elasticsearch/volatile/logs:rw",
		"--tmpfs", "/tmp:rw",
		"--ulimit", "memlock=-1:-1",
		"docker.elastic.co/elasticsearch/elasticsearch:8.17.1",
	}
	// "9200/tcp" refers to the default container port where elasticsearch server runs
	esHostQuery = `{{index .NetworkSettings.Ports "9200/tcp" 0 "HostIp"}}:{{index .NetworkSettings.Ports "9200/tcp" 0 "HostPort"}}`
)

// ElasticsearchTestServer is a single node elastic-search
// cluster running inside docker.
// use NewElasticsearchTestServer to instantiate the server
type ElasticsearchTestServer struct {
	url         *url.URL
	containerID string
	client      *elasticsearch.Client
}

// NewElasticsearchTestServer creates a new instance of elasticsearch test server.
// It runs a single node elasticsearch cluster in docker, exposing the REST
// API over a random ephemeral port.
// OR if the environment variable ES_TEST_SERVER_URL is set, it acts as
// a dumb proxy to it.
// The idea is to be able to easily run integration tests in local environments,
// while also being able to leverage a running ES instance for testing (for instance in CI pipelines)
// Make sure to call server.Close() once you're done, otherwise the docker
// container may be left running indefinitely in the background.
func NewElasticsearchTestServer() *ElasticsearchTestServer {
	var server ElasticsearchTestServer
	defer func() {
		if p := recover(); p != nil {
			server.Close()
			panic(p)
		}
	}()

	esURL, ok := os.LookupEnv("ES_TEST_SERVER_URL")
	if ok {
		// use TestServer as a proxy to an existing elasticsearch instance
		u, err := url.Parse(esURL)
		if err != nil {
			panic(fmt.Sprintf("error parsing elastisearch url: %v", err))
		}
		server.url = u
	} else {
		// run a new elasticsearch server inside a docker container
		idBytes, err := exec.Command(elasticSearchCmdLine[0], elasticSearchCmdLine[1:]...).Output()
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			if ok {
				err = fmt.Errorf("%v: %s", err, exitErr.Stderr)
			}
			panic(fmt.Sprintf("failed to start elasticsearch server: %v", err))
		}
		server.containerID = strings.TrimSpace(string(idBytes))

		// obtain the ephemeral host port which is bound to the container port
		hostBytes, err := exec.Command("docker", "inspect", "-f", esHostQuery, server.containerID).Output()
		if err != nil {
			panic(fmt.Sprintf("unable to obtain metadata for elasticsearch server: %v", err))
		}

		// add the server url to server
		server.url = &url.URL{
			Scheme: "http",
			Host:   strings.TrimSpace(string(hostBytes)),
		}
	}

	// wait for the elasticsearch server to come up
	timeout := 5 * time.Minute
	if err := server.wait4Ready(timeout); err != nil {
		panic(fmt.Sprintf("error checking elasticsearch status: %v", err))
	}

	// create the client
	var err error
	server.client, err = elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: []string{
				server.url.String(),
			},
			// uncomment below code to debug request and response to elasticsearch
			//Logger: &estransport.ColorLogger{
			//	Output:             os.Stdout,
			//	EnableRequestBody:  true,
			//	EnableResponseBody: true,
			//},
		},
	)
	if err != nil {
		panic(fmt.Sprintf("error creating elasticsearch client: %v", err))
	}

	return &server
}

// NewClient returns an elasticsearch client for the test server
// Calling this method issues a DELETE /_all call to the elasticsearch
// server, effectively resetting it.
func (srv *ElasticsearchTestServer) NewClient() (*elasticsearch.Client, error) {
	if err := srv.purge(srv.client); err != nil {
		return nil, fmt.Errorf("error purging elasticsearch: %w", err)
	}
	return srv.client, nil
}

func (srv *ElasticsearchTestServer) Close() error {
	if strings.TrimSpace(srv.containerID) == "" {
		return nil
	}
	return exec.Command("docker", "kill", srv.containerID).Run()
}

func (srv *ElasticsearchTestServer) purge(cli *elasticsearch.Client) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("purge: %w", err)
		}
	}()

	// List all non-system indices first. ES 8.x protects system indices
	// (dot-prefixed) from deletion, so we must discover user indices
	// explicitly and delete only those.
	catReq, err := http.NewRequest("GET", "/_cat/indices?h=index&format=json&expand_wildcards=open,closed", nil)
	if err != nil {
		return err
	}
	catRes, err := cli.Perform(catReq)
	if err != nil {
		return err
	}
	defer catRes.Body.Close()
	body, err := io.ReadAll(catRes.Body)
	if err != nil {
		return err
	}

	// Extract index names from the JSON array of {"index":"name"} objects.
	// Use simple string parsing to avoid adding an encoding/json dependency
	// just for tests. The response looks like: [{"index":"foo"},{"index":"bar"}]
	var indices []string
	for _, part := range strings.Split(string(body), "\"index\":\"") {
		if idx := strings.Index(part, "\""); idx > 0 {
			name := part[:idx]
			// Skip system indices (dot-prefixed)
			if !strings.HasPrefix(name, ".") {
				indices = append(indices, name)
			}
		}
	}

	// Nothing to delete on a fresh ES instance.
	if len(indices) == 0 {
		return nil
	}

	target := strings.Join(indices, ",")
	delReq, err := http.NewRequest("DELETE", "/"+target, nil)
	if err != nil {
		return err
	}
	q := delReq.URL.Query()
	q.Set("ignore_unavailable", "true")
	delReq.URL.RawQuery = q.Encode()
	delRes, err := cli.Perform(delReq)
	if err != nil {
		return err
	}
	defer delRes.Body.Close()
	if delRes.StatusCode > 299 {
		respBody, _ := io.ReadAll(delRes.Body)
		return fmt.Errorf("elasticsearch server returned status code %d: %s", delRes.StatusCode, respBody)
	}
	return nil
}

func (srv *ElasticsearchTestServer) wait4Ready(timeout time.Duration) error {
	catURL := srv.url.ResolveReference(&url.URL{Path: "/_cat"})
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		res, err := http.Get(catURL.String())
		if err != nil {
			continue
		}
		res.Body.Close()
		if res.StatusCode == 200 {
			return nil
		}
	}
	return fmt.Errorf("timed out after %s", timeout)
}
