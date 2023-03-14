package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/goto/compass/core/tag"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var assetID = uuid.NewString()
var sampleTag = tag.Tag{
	AssetID:             assetID,
	TemplateURN:         "governance_policy",
	TemplateDisplayName: "Governance Policy",
	TemplateDescription: "Template that is mandatory to be used.",
	TagValues: []tag.TagValue{
		{
			FieldID:          1,
			FieldValue:       "Public",
			FieldURN:         "classification",
			FieldDisplayName: "classification",
			FieldDescription: "The classification of this asset",
			FieldDataType:    "enumerated",
			FieldRequired:    true,
			FieldOptions:     []string{"Public", "Restricted"},
		},
		{
			FieldID:          2,
			FieldValue:       true,
			FieldURN:         "is_encrypted",
			FieldDisplayName: "Is Encrypted?",
			FieldDescription: "Specify whether this asset is encrypted or not.",
			FieldDataType:    "boolean",
			FieldRequired:    true,
		},
	},
}

var sampleTagPB = &compassv1beta1.Tag{
	AssetId:             assetID,
	TemplateUrn:         "governance_policy",
	TemplateDisplayName: "Governance Policy",
	TemplateDescription: "Template that is mandatory to be used.",
	TagValues: []*compassv1beta1.TagValue{
		{
			FieldId:          1,
			FieldValue:       structpb.NewStringValue("Public"),
			FieldUrn:         "classification",
			FieldDisplayName: "classification",
			FieldDescription: "The classification of this asset",
			FieldDataType:    "enumerated",
			FieldRequired:    true,
			FieldOptions:     []string{"Public", "Restricted"},
		},
		{
			FieldId:          2,
			FieldValue:       structpb.NewBoolValue(true),
			FieldUrn:         "is_encrypted",
			FieldDisplayName: "Is Encrypted?",
			FieldDescription: "Specify whether this asset is encrypted or not.",
			FieldDataType:    "boolean",
			FieldRequired:    true,
		},
	},
}

func TestGetTagByAssetAndTemplate(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetTagByAssetAndTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.GetTagByAssetAndTemplateResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if asset id is empty`,
			Request: &compassv1beta1.GetTagByAssetAndTemplateRequest{
				AssetId:     "",
				TemplateUrn: "sample-template",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.GetTagByAssetAndTemplateRequest{
				AssetId:     assetID,
				TemplateUrn: "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return not found if template does not exist`,
			Request: &compassv1beta1.GetTagByAssetAndTemplateRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().FindTagByAssetIDAndTemplateURN(ctx, assetID, sampleTemplate.URN).Return(tag.Tag{}, tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description: `should return not found if tag does not exist`,
			Request: &compassv1beta1.GetTagByAssetAndTemplateRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().FindTagByAssetIDAndTemplateURN(ctx, assetID, sampleTemplate.URN).Return(tag.Tag{}, tag.NotFoundError{
					AssetID:  assetID,
					Template: sampleTemplate.URN,
				})
			},
		},
		{
			Description: `should return internal server error if found unexpected error`,
			Request: &compassv1beta1.GetTagByAssetAndTemplateRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().FindTagByAssetIDAndTemplateURN(ctx, assetID, sampleTemplate.URN).Return(tag.Tag{}, errors.New("unexpected error"))
			},
		},
		{
			Description: `should return ok and tag`,
			Request: &compassv1beta1.GetTagByAssetAndTemplateRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().FindTagByAssetIDAndTemplateURN(ctx, assetID, sampleTemplate.URN).Return(sampleTag, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetTagByAssetAndTemplateResponse) error {
				var tagValuesPB []*compassv1beta1.TagValue
				for _, tv := range sampleTag.TagValues {
					tvPB, err := tagValueToProto(tv)
					if err != nil {
						return err
					}
					tagValuesPB = append(tagValuesPB, tvPB)
				}

				expected := &compassv1beta1.GetTagByAssetAndTemplateResponse{
					Data: &compassv1beta1.Tag{
						AssetId:             sampleTag.AssetID,
						TemplateUrn:         sampleTag.TemplateURN,
						TagValues:           tagValuesPB,
						TemplateDisplayName: sampleTag.TemplateDisplayName,
						TemplateDescription: sampleTag.TemplateDescription,
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockTagSvc := new(mocks.TagService)
			mockTemplateSvc := new(mocks.TagTemplateService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagSvc, mockTemplateSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockTagSvc.AssertExpectations(t)
			defer mockTemplateSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, nil, nil, mockTagSvc, mockTemplateSvc, mockUserSvc)

			got, err := handler.GetTagByAssetAndTemplate(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestCreateTagAsset(t *testing.T) {
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.CreateTagAssetRequest{
			AssetId:             sampleTagPB.GetAssetId(),
			TemplateUrn:         sampleTagPB.GetTemplateUrn(),
			TagValues:           sampleTagPB.TagValues,
			TemplateDisplayName: sampleTagPB.TemplateDisplayName,
			TemplateDescription: sampleTagPB.TemplateDescription,
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateTagAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.CreateTagAssetResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if asset id is empty`,
			Request: &compassv1beta1.CreateTagAssetRequest{
				AssetId:     "",
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.CreateTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: "",
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if tag values is empty`,
			Request: &compassv1beta1.CreateTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return not found if template does not exist`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().CreateTag(ctx, &sampleTag).Return(tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return invalid argument if there is validation error`,
			Request:      validRequest,
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().CreateTag(ctx, &sampleTag).Return(tag.ValidationError{Err: errors.New("validation error")})
			},
		},
		{
			Description:  `should return internal server error if found error during insert`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().CreateTag(ctx, &sampleTag).Return(errors.New("unexpected error during insert"))
			},
		},
		{
			Description:  `should return already exist if found duplicated asset during insert`,
			Request:      validRequest,
			ExpectStatus: codes.AlreadyExists,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().CreateTag(ctx, &sampleTag).Return(tag.DuplicateError{})
			},
		},
		{
			Description:  `should return ok and domain is inserted if found no error`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().CreateTag(ctx, &sampleTag).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.CreateTagAssetResponse) error {
				expected := &compassv1beta1.CreateTagAssetResponse{
					Data: sampleTagPB,
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockTagSvc := new(mocks.TagService)
			mockTemplateSvc := new(mocks.TagTemplateService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagSvc, mockTemplateSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockTagSvc.AssertExpectations(t)
			defer mockTemplateSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, nil, nil, mockTagSvc, mockTemplateSvc, mockUserSvc)

			got, err := handler.CreateTagAsset(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestUpdateTagAsset(t *testing.T) {
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.UpdateTagAssetRequest{
			AssetId:             sampleTagPB.GetAssetId(),
			TemplateUrn:         sampleTagPB.GetTemplateUrn(),
			TagValues:           sampleTagPB.TagValues,
			TemplateDisplayName: sampleTagPB.TemplateDisplayName,
			TemplateDescription: sampleTagPB.TemplateDescription,
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpdateTagAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.UpdateTagAssetResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if asset id is empty`,
			Request: &compassv1beta1.UpdateTagAssetRequest{
				AssetId:     "",
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.UpdateTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: "",
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if tag values is empty`,
			Request: &compassv1beta1.UpdateTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return not found if tag could not be found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().UpdateTag(ctx, &sampleTag).Return(tag.NotFoundError{})
			},
		},
		{
			Description:  `should return internal server error if found error during update`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().UpdateTag(ctx, &sampleTag).Return(errors.New("unexpected error during update"))
			},
		},
		{
			Description:  `should return ok and domain is updated if found no error`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().UpdateTag(ctx, &sampleTag).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.UpdateTagAssetResponse) error {
				expected := &compassv1beta1.UpdateTagAssetResponse{
					Data: sampleTagPB,
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockTagSvc := new(mocks.TagService)
			mockTemplateSvc := new(mocks.TagTemplateService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagSvc, mockTemplateSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockTagSvc.AssertExpectations(t)
			defer mockTemplateSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, nil, nil, mockTagSvc, mockTemplateSvc, mockUserSvc)

			got, err := handler.UpdateTagAsset(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestDeleteTagAsset(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.DeleteTagAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if asset id is empty`,
			Request: &compassv1beta1.DeleteTagAssetRequest{
				AssetId:     "",
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.DeleteTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return not found if template does not exist`,
			Request: &compassv1beta1.DeleteTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().DeleteTag(ctx, assetID, sampleTagPB.TemplateUrn).Return(tag.TemplateNotFoundError{})
			},
		},
		{
			Description: `should return internal server error found unexpected error`,
			Request: &compassv1beta1.DeleteTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().DeleteTag(ctx, assetID, sampleTagPB.TemplateUrn).Return(errors.New("unexpected error"))
			},
		},
		{
			Description: `should return ok if delete success`,
			Request: &compassv1beta1.DeleteTagAssetRequest{
				AssetId:     assetID,
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().DeleteTag(ctx, assetID, sampleTagPB.TemplateUrn).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockTagSvc := new(mocks.TagService)
			mockTemplateSvc := new(mocks.TagTemplateService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagSvc, mockTemplateSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockTagSvc.AssertExpectations(t)
			defer mockTemplateSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, nil, nil, mockTagSvc, mockTemplateSvc, mockUserSvc)

			_, err := handler.DeleteTagAsset(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestGetAllTagsByAsset(t *testing.T) {
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.GetAllTagsByAssetRequest{
			AssetId: assetID,
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllTagsByAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.GetAllTagsByAssetResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if asset id is empty`,
			Request: &compassv1beta1.GetAllTagsByAssetRequest{
				AssetId: "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return internal server error if found unexpected error`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().GetTagsByAssetID(ctx, sampleTagPB.AssetId).Return(nil, errors.New("unexpected error"))
			},
		},
		{
			Description:  `should return ok and tags for the specified asset`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				ts.EXPECT().GetTagsByAssetID(ctx, sampleTagPB.AssetId).Return([]tag.Tag{sampleTag}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllTagsByAssetResponse) error {
				expected := &compassv1beta1.GetAllTagsByAssetResponse{
					Data: []*compassv1beta1.Tag{sampleTagPB},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockTagSvc := new(mocks.TagService)
			mockTemplateSvc := new(mocks.TagTemplateService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagSvc, mockTemplateSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockTagSvc.AssertExpectations(t)
			defer mockTemplateSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, nil, nil, mockTagSvc, mockTemplateSvc, mockUserSvc)

			got, err := handler.GetAllTagsByAsset(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestTagToProto(t *testing.T) {
	type testCase struct {
		Title       string
		Tag         tag.Tag
		ExpectProto *compassv1beta1.Tag
	}

	var testCases = []testCase{
		{
			Title:       "should return empty field value pb if tag values is empty",
			Tag:         tag.Tag{AssetID: "1111-2222-3333-4444"},
			ExpectProto: &compassv1beta1.Tag{AssetId: "1111-2222-3333-4444"},
		},
		{
			Title:       "should return tag value pb if tag values is not empty",
			Tag:         tag.Tag{AssetID: "1111-2222-3333-4444", TagValues: []tag.TagValue{{FieldID: 123, FieldURN: "urn"}}},
			ExpectProto: &compassv1beta1.Tag{AssetId: "1111-2222-3333-4444", TagValues: []*compassv1beta1.TagValue{{FieldId: 123, FieldUrn: "urn"}}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got, err := tagToProto(tc.Tag)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestTagFromProto(t *testing.T) {
	type testCase struct {
		Title  string
		PB     *compassv1beta1.Tag
		Expect tag.Tag
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty tag values if tag values pb are not empty",
			PB:     &compassv1beta1.Tag{AssetId: "1111-2222-3333-4444", TagValues: []*compassv1beta1.TagValue{{FieldId: 123, FieldUrn: "urn"}}},
			Expect: tag.Tag{AssetID: "1111-2222-3333-4444", TagValues: []tag.TagValue{{FieldID: 123, FieldURN: "urn"}}},
		},
		{
			Title:  "should return empty tag values if tag values pb are empty",
			PB:     &compassv1beta1.Tag{AssetId: "1111-2222-3333-4444"},
			Expect: tag.Tag{AssetID: "1111-2222-3333-4444"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tagFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}

func TestTagValueToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		TagValue    tag.TagValue
		ExpectProto *compassv1beta1.TagValue
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb and empty field value pb if timestamp and field value are empty or zero",
			TagValue:    tag.TagValue{FieldID: 123, FieldURN: "urn"},
			ExpectProto: &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn"},
		},
		{
			Title:       "should return timestamp pb and field value pb if timestamp and field value are not empty or zero",
			TagValue:    tag.TagValue{FieldID: 123, FieldURN: "urn", FieldValue: "a value", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn", FieldValue: structpb.NewStringValue("a value"), CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got, err := tagValueToProto(tc.TagValue)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestTagValueFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.TagValue
		Expect tag.TagValue
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty time.Time and field value if timestamp pb and field value pb are not empty or zero",
			PB:     &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn", FieldValue: structpb.NewStringValue("a value"), CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: tag.TagValue{FieldID: 123, FieldURN: "urn", FieldValue: "a value", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time and empty field value if timestamp pb and field value pb are empty or zero",
			PB:     &compassv1beta1.TagValue{FieldId: 123, FieldUrn: "urn"},
			Expect: tag.TagValue{FieldID: 123, FieldURN: "urn"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := tagValueFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
