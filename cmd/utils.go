package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

func parseFile(filePath string, v interface{}) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := json.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
	default:
		return errors.New("unsupported file type")
	}

	return nil
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func setCtxHeader(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{headerKey: headerValue})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

func makeMapFromString(commaSepStr string) map[string]string {
	m := make(map[string]string)
	keyValArray := strings.Split(commaSepStr, ",")
	for _, s := range keyValArray {
		arr := strings.Split(s, ":")
		m[arr[0]] = arr[1]
	}
	return m
}
