package star

import (
	"testing"

	"github.com/odpf/columbus/asset"
	"gotest.tools/assert"
)

func TestValidate(t *testing.T) {
	type testCase struct {
		Title       string
		Star        *Star
		ExpectError error
	}

	var testCases = []testCase{
		{
			Title:       "should return error invalid if user is nil",
			Star:        nil,
			ExpectError: InvalidError{},
		},
		{
			Title:       "should return error invalid if assert urn is empty",
			Star:        &Star{Asset: asset.Asset{Type: "asset-type"}},
			ExpectError: InvalidError{AssetType: "asset-type"},
		},
		{
			Title:       "should return error invalid if assert type is empty",
			Star:        &Star{Asset: asset.Asset{URN: "asset-urn"}},
			ExpectError: InvalidError{AssetURN: "asset-urn"},
		},
		{
			Title:       "should return nil if star is valid",
			Star:        &Star{Asset: asset.Asset{URN: "asset-urn", Type: "asset-type"}},
			ExpectError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {

			err := testCase.Star.ValidateAssetURN()
			assert.Equal(t, testCase.ExpectError, err)
		})
	}
}
