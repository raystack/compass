package elasticsearch_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/internal/store/elasticsearch/testutil"
)

var daggerType = asset.Type("dagger")

var esTestServer *testutil.ElasticsearchTestServer

func TestMain(m *testing.M) {
	// TODO(Aman): this block makes it impossible to skip starting
	// an elasticsearch server. That means you can't run unit tests
	// standalone :/
	esTestServer = testutil.NewElasticsearchTestServer()

	exitCode := m.Run()

	if err := esTestServer.Close(); err != nil {
		fmt.Println("Error closing elasticsearch test server:", err)
		return
	}
	os.Exit(exitCode)
}
