package store_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/odpf/columbus/models"
	"github.com/odpf/columbus/store/testutil"
)

var daggerType = models.Type{
	Name:           "dagger",
	Classification: models.TypeClassificationResource,
	Fields: models.TypeFields{
		ID:     "urn",
		Labels: []string{"country"},
		Title:  "title",
	},
}

var esTestServer *testutil.ElasticsearchTestServer

func TestMain(m *testing.M) {
	// TODO(Aman): this block makes it impossible to skip starting
	// an elasticsearch server. That means you can't run unit tests
	// standlone :/
	esTestServer = testutil.NewElasticsearchTestServer()
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
