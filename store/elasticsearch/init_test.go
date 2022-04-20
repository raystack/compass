package elasticsearch_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/store/elasticsearch/testutil"
)

var daggerType = asset.Type("dagger")

var esTestServer *testutil.ElasticsearchTestServer

func TestMain(m *testing.M) {
	// TODO(Aman): this block makes it impossible to skip starting
	// an elasticsearch server. That means you can't run unit tests
	// standlone :/
	esTestServer = testutil.NewElasticsearchTestServer()
	exitCode := m.Run()

	if err := esTestServer.Close(); err != nil {
		fmt.Println("Error closing elasticsearch test server:", err)
		return
	}
	os.Exit(exitCode)
}

// name this somethings that's more generic
func incorrectResultsError(expect, actual interface{}) error {
	out := new(bytes.Buffer)
	out.WriteString("\n=== Expected ===\n")
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(expect)
	if err != nil {
		panic(err)
	}
	out.WriteString("=== Actual ===\n")
	err = encoder.Encode(actual)
	if err != nil {
		panic(err)
	}
	return errors.New(out.String())
}
