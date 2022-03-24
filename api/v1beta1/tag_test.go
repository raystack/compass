package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/columbus/api"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

var sampleTag = tag.Tag{
	RecordURN:           "sample-urn",
	RecordType:          "sample-type",
	TemplateURN:         "governance_policy",
	TemplateDisplayName: "Governance Policy",
	TemplateDescription: "Template that is mandatory to be used.",
	TagValues: []tag.TagValue{
		{
			FieldID:          1,
			FieldValue:       "Public",
			FieldURN:         "classification",
			FieldDisplayName: "classification",
			FieldDescription: "The classification of this record",
			FieldDataType:    "enumerated",
			FieldRequired:    true,
			FieldOptions:     []string{"Public", "Restricted"},
		},
		{
			FieldID:          2,
			FieldValue:       true,
			FieldURN:         "is_encrypted",
			FieldDisplayName: "Is Encrypted?",
			FieldDescription: "Specify whether this record is encrypted or not.",
			FieldDataType:    "boolean",
			FieldRequired:    true,
		},
	},
}

var sampleTagPB = &compassv1beta1.Tag{
	RecordUrn:           "sample-urn",
	RecordType:          "sample-type",
	TemplateUrn:         "governance_policy",
	TemplateDisplayName: "Governance Policy",
	TemplateDescription: "Template that is mandatory to be used.",
	TagValues: []*compassv1beta1.TagValue{
		{
			FieldId:          1,
			FieldValue:       structpb.NewStringValue("Public"),
			FieldUrn:         "classification",
			FieldDisplayName: "classification",
			FieldDescription: "The classification of this record",
			FieldDataType:    "enumerated",
			FieldRequired:    true,
			FieldOptions:     []string{"Public", "Restricted"},
		},
		{
			FieldId:          2,
			FieldValue:       structpb.NewBoolValue(true),
			FieldUrn:         "is_encrypted",
			FieldDisplayName: "Is Encrypted?",
			FieldDescription: "Specify whether this record is encrypted or not.",
			FieldDataType:    "boolean",
			FieldRequired:    true,
		},
	},
}

func TestGetTagsByRecordAndTemplate(t *testing.T) {
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetTagsByRecordAndTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.GetTagsByRecordAndTemplateResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if record urn is empty`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "sample-type",
				RecordUrn:   "",
				TemplateUrn: "sample-template",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if type is empty`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "",
				RecordUrn:   "sample-urn",
				TemplateUrn: "sample-template",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "sample-type",
				RecordUrn:   "sample-urn",
				TemplateUrn: "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return not found if template does not exist`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "sample-type",
				RecordUrn:   "sample-urn",
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{}, tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description: `should return not found if tag does not exist`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "sample-type",
				RecordUrn:   "sample-urn",
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType:  "sample-type",
					RecordURN:   "sample-urn",
					TemplateURN: sampleTemplate.URN,
				}).Return(nil, tag.NotFoundError{
					URN:      "sample-urn",
					Type:     "sample-type",
					Template: sampleTemplate.URN,
				})
			},
		},
		{
			Description: `should return internal server error if found unexpected error`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "sample-type",
				RecordUrn:   "sample-urn",
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType:  "sample-type",
					RecordURN:   "sample-urn",
					TemplateURN: sampleTemplate.URN,
				}).Return(nil, errors.New("unexpected error"))
			},
		},
		{
			Description: `should return ok and tag`,
			Request: &compassv1beta1.GetTagsByRecordAndTemplateRequest{
				Type:        "sample-type",
				RecordUrn:   "sample-urn",
				TemplateUrn: sampleTemplate.URN,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType:  "sample-type",
					RecordURN:   "sample-urn",
					TemplateURN: sampleTemplate.URN,
				}).Return([]tag.Tag{sampleTag}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetTagsByRecordAndTemplateResponse) error {
				var tagValuesPB []*compassv1beta1.TagValue
				for _, tv := range sampleTag.TagValues {
					tvPB, err := tv.ToProto()
					if err != nil {
						return err
					}
					tagValuesPB = append(tagValuesPB, tvPB)
				}

				expected := &compassv1beta1.GetTagsByRecordAndTemplateResponse{
					Data: &compassv1beta1.Tag{
						RecordType:          sampleTag.RecordType,
						RecordUrn:           sampleTag.RecordURN,
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
			ctx := context.Background()
			logger := log.NewNoop()
			mockTagRepo := new(mocks.TagRepository)
			mockTemplateRepo := new(mocks.TagTemplateRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagRepo, mockTemplateRepo)
			}
			defer mockTagRepo.AssertExpectations(t)
			defer mockTemplateRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService: service,
			})

			got, err := handler.GetTagsByRecordAndTemplate(ctx, tc.Request)
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

func TestCreateTag(t *testing.T) {
	validRequest := &compassv1beta1.CreateTagRequest{
		RecordType:          sampleTagPB.GetRecordType(),
		RecordUrn:           sampleTagPB.GetRecordUrn(),
		TemplateUrn:         sampleTagPB.GetTemplateUrn(),
		TagValues:           sampleTagPB.TagValues,
		TemplateDisplayName: sampleTagPB.TemplateDisplayName,
		TemplateDescription: sampleTagPB.TemplateDescription,
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateTagRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.CreateTagResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if record type is empty`,
			Request: &compassv1beta1.CreateTagRequest{
				RecordType:  "",
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if record urn is empty`,
			Request: &compassv1beta1.CreateTagRequest{
				RecordType:  sampleTagPB.GetRecordType(),
				RecordUrn:   "",
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.CreateTagRequest{
				RecordType:  sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: "",
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if tag values is empty`,
			Request: &compassv1beta1.CreateTagRequest{
				RecordType:  sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return not found if template does not exist`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Create(ctx, &sampleTag).Return(tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return invalid argument if there is validation error`,
			Request:      validRequest,
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Create(ctx, &sampleTag).Return(tag.ValidationError{Err: errors.New("validation error")})
			},
		},
		{
			Description:  `should return internal server error if found error during insert`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Create(ctx, &sampleTag).Return(errors.New("unexpected error during insert"))
			},
		},
		{
			Description:  `should return already exist if found duplicated record during insert`,
			Request:      validRequest,
			ExpectStatus: codes.AlreadyExists,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Create(ctx, &sampleTag).Return(tag.DuplicateError{})
			},
		},
		{
			Description:  `should return ok and domain is inserted if found no error`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Create(ctx, &sampleTag).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.CreateTagResponse) error {
				expected := &compassv1beta1.CreateTagResponse{
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
			ctx := context.Background()
			logger := log.NewNoop()
			mockTagRepo := new(mocks.TagRepository)
			mockTemplateRepo := new(mocks.TagTemplateRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagRepo, mockTemplateRepo)
			}
			defer mockTagRepo.AssertExpectations(t)
			defer mockTemplateRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService: service,
			})

			got, err := handler.CreateTag(ctx, tc.Request)
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

func TestUpdateTag(t *testing.T) {
	validRequest := &compassv1beta1.UpdateTagRequest{
		Type:                sampleTagPB.GetRecordType(),
		RecordUrn:           sampleTagPB.GetRecordUrn(),
		TemplateUrn:         sampleTagPB.GetTemplateUrn(),
		TagValues:           sampleTagPB.TagValues,
		TemplateDisplayName: sampleTagPB.TemplateDisplayName,
		TemplateDescription: sampleTagPB.TemplateDescription,
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpdateTagRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.UpdateTagResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if type is empty`,
			Request: &compassv1beta1.UpdateTagRequest{
				Type:        "",
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if record urn is empty`,
			Request: &compassv1beta1.UpdateTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   "",
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.UpdateTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: "",
				TagValues:   sampleTagPB.TagValues,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if tag values is empty`,
			Request: &compassv1beta1.UpdateTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return not found if tag could not be found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType:  sampleTagPB.RecordType,
					RecordURN:   sampleTagPB.RecordUrn,
					TemplateURN: sampleTagPB.TemplateUrn,
				}).Return([]tag.Tag{}, nil)
			},
		},
		{
			Description:  `should return internal server error if found error during update`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType:  sampleTagPB.RecordType,
					RecordURN:   sampleTagPB.RecordUrn,
					TemplateURN: sampleTagPB.TemplateUrn,
				}).Return([]tag.Tag{sampleTag}, nil)
				tr.EXPECT().Update(ctx, &sampleTag).Return(errors.New("unexpected error during update"))
			},
		},
		{
			Description:  `should return ok and domain is updated if found no error`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType:  sampleTagPB.RecordType,
					RecordURN:   sampleTagPB.RecordUrn,
					TemplateURN: sampleTagPB.TemplateUrn,
				}).Return([]tag.Tag{sampleTag}, nil)
				tr.EXPECT().Update(ctx, &sampleTag).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.UpdateTagResponse) error {
				expected := &compassv1beta1.UpdateTagResponse{
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
			ctx := context.Background()
			logger := log.NewNoop()
			mockTagRepo := new(mocks.TagRepository)
			mockTemplateRepo := new(mocks.TagTemplateRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagRepo, mockTemplateRepo)
			}
			defer mockTagRepo.AssertExpectations(t)
			defer mockTemplateRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService: service,
			})

			got, err := handler.UpdateTag(ctx, tc.Request)
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

func TestDeleteTag(t *testing.T) {
	type testCase struct {
		Description  string
		Request      *compassv1beta1.DeleteTagRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if type is empty`,
			Request: &compassv1beta1.DeleteTagRequest{
				Type:        "",
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if record urn is empty`,
			Request: &compassv1beta1.DeleteTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   "",
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if template urn is empty`,
			Request: &compassv1beta1.DeleteTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return not found if template does not exist`,
			Request: &compassv1beta1.DeleteTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{{}}, nil)
				tr.EXPECT().Delete(ctx, tag.Tag{
					RecordType:  sampleTagPB.RecordType,
					RecordURN:   sampleTagPB.RecordUrn,
					TemplateURN: sampleTagPB.TemplateUrn,
				}).Return(tag.TemplateNotFoundError{})
			},
		},
		{
			Description: `should return internal server error found unexpected error`,
			Request: &compassv1beta1.DeleteTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{{}}, nil)
				tr.EXPECT().Delete(ctx, tag.Tag{
					RecordType:  sampleTagPB.RecordType,
					RecordURN:   sampleTagPB.RecordUrn,
					TemplateURN: sampleTagPB.TemplateUrn,
				}).Return(errors.New("unexpected error"))
			},
		},
		{
			Description: `should return ok if delete success`,
			Request: &compassv1beta1.DeleteTagRequest{
				Type:        sampleTagPB.GetRecordType(),
				RecordUrn:   sampleTagPB.GetRecordUrn(),
				TemplateUrn: sampleTagPB.GetTemplateUrn(),
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{{}}, nil)
				tr.EXPECT().Delete(ctx, tag.Tag{
					RecordType:  sampleTagPB.RecordType,
					RecordURN:   sampleTagPB.RecordUrn,
					TemplateURN: sampleTagPB.TemplateUrn,
				}).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()
			logger := log.NewNoop()
			mockTagRepo := new(mocks.TagRepository)
			mockTemplateRepo := new(mocks.TagTemplateRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagRepo, mockTemplateRepo)
			}
			defer mockTagRepo.AssertExpectations(t)
			defer mockTemplateRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService: service,
			})

			_, err := handler.DeleteTag(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestGetTagsByRecord(t *testing.T) {
	validRequest := &compassv1beta1.GetTagsByRecordRequest{
		Type:      sampleTagPB.GetRecordType(),
		RecordUrn: sampleTagPB.GetRecordUrn(),
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetTagsByRecordRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.GetTagsByRecordResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return invalid argument if type is empty`,
			Request: &compassv1beta1.GetTagsByRecordRequest{
				Type:      "",
				RecordUrn: sampleTagPB.GetRecordUrn(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if record urn is empty`,
			Request: &compassv1beta1.GetTagsByRecordRequest{
				Type:      sampleTagPB.GetRecordType(),
				RecordUrn: "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return internal server error if found unexpected error`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType: sampleTagPB.RecordType,
					RecordURN:  sampleTagPB.RecordUrn,
				}).Return(nil, errors.New("unexpected error"))
			},
		},
		{
			Description:  `should return ok and tags for the specified record`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				tr.EXPECT().Read(ctx, tag.Tag{
					RecordType: sampleTagPB.RecordType,
					RecordURN:  sampleTagPB.RecordUrn,
				}).Return([]tag.Tag{sampleTag}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetTagsByRecordResponse) error {
				expected := &compassv1beta1.GetTagsByRecordResponse{
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
			ctx := context.Background()
			logger := log.NewNoop()
			mockTagRepo := new(mocks.TagRepository)
			mockTemplateRepo := new(mocks.TagTemplateRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockTagRepo, mockTemplateRepo)
			}
			defer mockTagRepo.AssertExpectations(t)
			defer mockTemplateRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService: service,
			})

			got, err := handler.GetTagsByRecord(ctx, tc.Request)
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
