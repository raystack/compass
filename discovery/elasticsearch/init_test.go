package elasticsearch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"testing"

	store "github.com/odpf/columbus/discovery/elasticsearch"
	"github.com/odpf/columbus/discovery/elasticsearch/testutil"
)

var esTestServer *testutil.ElasticsearchTestServer

func TestMain(m *testing.M) {
	// TODO(Aman): this block makes it impossible to skip starting
	// an elasticsearch server. That means you can't run unit tests
	// standlone :/
	esTestServer = testutil.NewElasticsearchTestServer()
	if err := store.Migrate(context.TODO(), esTestServer.NewClient()); err != nil {
		log.Fatal(err)
	}
	defer esTestServer.Close()
	os.Exit(m.Run())
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
