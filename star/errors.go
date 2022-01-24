package star

import (
	"fmt"
	"strings"
)

type NotFoundError struct {
	AssetID   string
	AssetURN  string
	AssetType string
	UserID    string
}

func (e NotFoundError) Error() string {
	fields := []string{"could not find starred asset"}
	if e.AssetID != "" {
		fields = append(fields, fmt.Sprintf("with asset id \"%s\"", e.AssetID))
	}
	if e.AssetURN != "" {
		fields = append(fields, fmt.Sprintf("with asset urn \"%s\"", e.AssetURN))
	}
	if e.AssetType != "" {
		fields = append(fields, fmt.Sprintf("with asset type \"%s\"", e.AssetType))
	}
	if e.UserID != "" {
		fields = append(fields, fmt.Sprintf("by user id \"%s\"", e.UserID))
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

type UserNotFoundError struct {
	UserID string
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("could not find user with id \"%s\"", e.UserID)
}

type DuplicateRecordError struct {
	UserID  string
	AssetID string
}

func (e DuplicateRecordError) Error() string {
	return fmt.Sprintf("duplicate starred asset id \"%s\" with user id \"%s\"", e.AssetID, e.UserID)
}

type InvalidError struct {
	UserID    string
	AssetID   string
	AssetType string
	AssetURN  string
}

func (e InvalidError) Error() string {
	fields := []string{"empty input field"}
	if e.UserID != "" {
		fields = append(fields, fmt.Sprintf("with user id \"%s\"", e.UserID))
	}
	if e.AssetID != "" {
		fields = append(fields, fmt.Sprintf("with asset id \"%s\"", e.AssetID))
	}
	if e.AssetURN != "" {
		fields = append(fields, fmt.Sprintf("with asset urn \"%s\"", e.AssetURN))
	}
	if e.AssetType != "" {
		fields = append(fields, fmt.Sprintf("with asset type \"%s\"", e.AssetType))
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, " "))
}
