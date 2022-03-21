package asset

//go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --with-expecter --structname AssetRepository --filename asset_repository.go
import (
	"context"
	"fmt"
	"time"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/user"
	"github.com/r3labs/diff/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const MAX_ARRAY_FIELD_NUM = 10

type Repository interface {
	GetAll(context.Context, Config) ([]Asset, error)
	GetCount(context.Context, Config) (int, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	Find(ctx context.Context, urn string, typ Type, service string) (Asset, error)
	GetVersionHistory(ctx context.Context, cfg Config, id string) ([]Asset, error)
	GetByVersion(ctx context.Context, id string, version string) (Asset, error)
	Upsert(ctx context.Context, ast *Asset) (string, error)
	Delete(ctx context.Context, id string) error
}

// Asset is a model that wraps arbitrary data with Columbus' context
type Asset struct {
	ID          string                 `json:"id" diff:"-"`
	URN         string                 `json:"urn" diff:"-"`
	Type        Type                   `json:"type" diff:"-"`
	Service     string                 `json:"service" diff:"-"`
	Name        string                 `json:"name" diff:"name"`
	Description string                 `json:"description" diff:"description"`
	Data        map[string]interface{} `json:"data" diff:"data"`
	Labels      map[string]string      `json:"labels" diff:"labels"`
	Owners      []user.User            `json:"owners,omitempty" diff:"owners"`
	CreatedAt   time.Time              `json:"created_at" diff:"-"`
	UpdatedAt   time.Time              `json:"updated_at" diff:"-"`
	Version     string                 `json:"version" diff:"-"`
	UpdatedBy   user.User              `json:"updated_by" diff:"-"`
	Changelog   diff.Changelog         `json:"changelog,omitempty" diff:"-"`
}

// ToProto transforms struct to proto
func (a Asset) ToProto() (assetPB *compassv1beta1.Asset, err error) {
	var data *structpb.Struct
	if len(a.Data) > 0 {
		data, err = structpb.NewStruct(a.Data)
		if err != nil {
			return
		}
	}

	var labels *structpb.Struct
	if len(a.Labels) > 0 {
		labelsMapInterface := make(map[string]interface{}, len(a.Labels))
		for k, v := range a.Labels {
			labelsMapInterface[k] = v
		}
		labels, err = structpb.NewStruct(labelsMapInterface)
		if err != nil {
			return
		}
	}

	owners := []*compassv1beta1.User{}
	for _, o := range a.Owners {
		owners = append(owners, o.ToProto())
	}

	changelogProto, err := changelogToProto(a.Changelog)
	if err != nil {
		return nil, err
	}

	var createdAtPB *timestamppb.Timestamp
	if !a.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(a.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !a.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(a.UpdatedAt)
	}

	assetPB = &compassv1beta1.Asset{
		Id:          a.ID,
		Urn:         a.URN,
		Type:        string(a.Type),
		Service:     a.Service,
		Name:        a.Name,
		Description: a.Description,
		Data:        data,
		Labels:      labels,
		Owners:      owners,
		Version:     a.Version,
		UpdatedBy:   a.UpdatedBy.ToProto(),
		Changelog:   changelogProto,
		CreatedAt:   createdAtPB,
		UpdatedAt:   updatedAtPB,
	}
	return
}

// NewFromProto transforms proto to struct
// changelog is not populated by user, it should always be processed and coming from the server
func NewFromProto(pb *compassv1beta1.Asset) Asset {
	var assetOwners []user.User
	for _, op := range pb.GetOwners() {
		if op == nil {
			continue
		}
		assetOwners = append(assetOwners, user.NewFromProto(op))
	}

	var labels map[string]string
	if pb.GetLabels() != nil {
		labels = make(map[string]string)
		for key, value := range pb.GetLabels().AsMap() {
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", value)
			labels[strKey] = strValue
		}
	}

	var dataValue map[string]interface{}
	if pb.GetData() != nil {
		dataValue = pb.GetData().AsMap()
	}

	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var updatedBy user.User
	if pb.GetUpdatedBy() != nil {
		updatedBy = user.NewFromProto(pb.GetUpdatedBy())
	}

	var clog diff.Changelog
	if pb.GetChangelog() != nil {
		for _, cg := range pb.GetChangelog().GetChanges() {
			if cg == nil {
				continue
			}
			clog = append(clog, newDiffChangeFromProto(cg))
		}
	}

	return Asset{
		ID:          pb.GetId(),
		URN:         pb.GetUrn(),
		Type:        Type(pb.GetType()),
		Service:     pb.GetService(),
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
		Data:        dataValue,
		Labels:      labels,
		Owners:      assetOwners,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Version:     pb.GetVersion(),
		Changelog:   clog,
		UpdatedBy:   updatedBy,
	}
}

// AssignDataFromProto populates asset.Data from *structpb.Struct data
func (a *Asset) AssignDataFromProto(pb *structpb.Struct) {
	if pb != nil {
		a.Data = pb.AsMap()
	}
}

// AssignLabelsFromProto populates asset.Labels from *structpb.Struct data
func (a *Asset) AssignLabelsFromProto(pb *structpb.Struct) {
	if pb != nil {
		pbMap := pb.AsMap()
		if len(pbMap) > 0 {
			a.Labels = make(map[string]string)
			for key, value := range pb.AsMap() {
				strKey := fmt.Sprintf("%v", key)
				strValue := fmt.Sprintf("%v", value)
				a.Labels[strKey] = strValue
			}
		}
	}
}

// Diff returns nil changelog with nil error if equal
// returns wrapped r3labs/diff Changelog struct with nil error if not equal
func (a *Asset) Diff(otherAsset *Asset) (diff.Changelog, error) {
	return diff.Diff(a, otherAsset, diff.DiscardComplexOrigin(), diff.AllowTypeMismatch(true))
}

// Patch appends asset with data from map. It mutates the asset itself.
// It is using json annotation of the struct to patch the correct keys
func (a *Asset) Patch(patchData map[string]interface{}) {
	patchAsset(a, patchData)
}

// ValidateConstraint checks whether non empty/nil fields fulfill the contract
func (a *Asset) ValidateConstraint() error {
	//if len(strings.TrimSpace(a.Type.String())) > 0 && !IsTypeStringValid(a.Type.String()) {
	//	return ErrUnknownType
	//}

	//if a.Service != nil && len(a.Service) > MAX_ARRAY_FIELD_NUM {
	//	return fmt.Errorf("service cannot be more than %d", MAX_ARRAY_FIELD_NUM)
	//}
	//
	//if a.Type != nil && len(a.Type) > MAX_ARRAY_FIELD_NUM {
	//	return fmt.Errorf("type cannot be more than %d", MAX_ARRAY_FIELD_NUM)
	//}

	return nil
}
