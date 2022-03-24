package asset

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/user"
	"github.com/r3labs/diff/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const BaseVersion = "0.1"

// AssetVersion is the changes summary of asset versions
type AssetVersion struct {
	ID        string         `json:"id" db:"id"`
	URN       string         `json:"urn" db:"urn"`
	Type      string         `json:"type" db:"type"`
	Service   string         `json:"service" db:"service"`
	Version   string         `json:"version" db:"version"`
	Changelog diff.Changelog `json:"changelog" db:"changelog"`
	UpdatedBy user.User      `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// ToProto transforms struct to proto
func (av AssetVersion) ToProto() (*compassv1beta1.Asset, error) {

	changelogProto, err := changelogToProto(av.Changelog)
	if err != nil {
		return nil, err
	}

	var createdAtPB *timestamppb.Timestamp
	if !av.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(av.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !av.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(av.UpdatedAt)
	}

	return &compassv1beta1.Asset{
		Id:        av.ID,
		Urn:       av.URN,
		Type:      string(av.Type),
		Service:   av.Service,
		Version:   av.Version,
		UpdatedBy: av.UpdatedBy.ToProto(),
		Changelog: changelogProto,
		CreatedAt: createdAtPB,
		UpdatedAt: updatedAtPB,
	}, nil
}

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
func changelogToProto(cl diff.Changelog) (*compassv1beta1.Changelog, error) {
	if len(cl) == 0 {
		return nil, nil
	}
	protoChanges := []*compassv1beta1.Change{}
	for _, ch := range cl {
		chProto, err := diffChangeToProto(ch)
		if err != nil {
			return nil, err
		}

		protoChanges = append(protoChanges, chProto)
	}
	return &compassv1beta1.Changelog{
		Changes: protoChanges,
	}, nil
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
