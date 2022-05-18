package asset

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/r3labs/diff/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

const BaseVersion = "0.1"

// ParseVersion returns error if version string is not in MAJOR.MINOR format
func ParseVersion(v string) (*semver.Version, error) {
	semverVersion, err := semver.NewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("invalid version \"%s\"", v)
	}
	return semverVersion, nil
}

// IncreaseMinorVersion bumps up the minor version +0.1
func IncreaseMinorVersion(v string) (string, error) {
	oldVersion, err := ParseVersion(v)
	if err != nil {
		return "", err
	}
	newVersion := oldVersion.IncMinor()
	return fmt.Sprintf("%d.%d", newVersion.Major(), newVersion.Minor()), nil
}

// changelogToProto transforms changelog struct to proto
func changelogToProto(cl diff.Changelog) ([]*compassv1beta1.Change, error) {
	if len(cl) == 0 {
		return nil, nil
	}
	var protoChanges []*compassv1beta1.Change
	for _, ch := range cl {
		chProto, err := diffChangeToProto(ch)
		if err != nil {
			return nil, err
		}

		protoChanges = append(protoChanges, chProto)
	}
	return protoChanges, nil
}

func diffChangeToProto(dc diff.Change) (*compassv1beta1.Change, error) {
	from, err := structpb.NewValue(dc.From)
	if err != nil {
		return nil, err
	}
	to, err := structpb.NewValue(dc.To)
	if err != nil {
		return nil, err
	}

	return &compassv1beta1.Change{
		Type: dc.Type,
		Path: dc.Path,
		From: from,
		To:   to,
	}, nil
}

// newDiffChangeFromProto converts Change proto to diff.Change
func newDiffChangeFromProto(pb *compassv1beta1.Change) diff.Change {
	var fromItf interface{}
	if pb.GetFrom() != nil {
		fromItf = pb.GetFrom().AsInterface()
	}

	var toItf interface{}
	if pb.GetTo() != nil {
		toItf = pb.GetTo().AsInterface()
	}

	return diff.Change{
		Type: pb.GetType(),
		Path: pb.GetPath(),
		From: fromItf,
		To:   toItf,
	}
}
