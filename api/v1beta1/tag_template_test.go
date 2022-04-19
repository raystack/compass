package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/compass/api"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/lib/mocks"
	"github.com/odpf/compass/tag"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

var sampleTemplate = tag.Template{
	URN:         "governance_policy",
	DisplayName: "Governance Policy",
	Description: "Template that is mandatory to be used.",
	Fields: []tag.Field{
		{
			ID:          1,
			URN:         "classification",
			DisplayName: "classification",
			Description: "The classification of this record",
			DataType:    "enumerated",
			Required:    true,
			Options:     []string{"Public", "Restricted"},
		},
		{
			ID:          2,
			URN:         "is_encrypted",
			DisplayName: "Is Encrypted?",
			Description: "Specify whether this record is encrypted or not.",
			DataType:    "boolean",
			Required:    true,
		},
	},
}

var sampleTemplatePB = &compassv1beta1.TagTemplate{
	Urn:         sampleTemplate.URN,
	DisplayName: sampleTemplate.DisplayName,
	Description: sampleTemplate.Description,
	Fields: []*compassv1beta1.TagTemplateField{
		{
			Id:          1,
			Urn:         "classification",
			DisplayName: "classification",
			Description: "The classification of this record",
			DataType:    "enumerated",
			Required:    true,
			Options:     []string{"Public", "Restricted"},
		},
		{
			Id:          2,
			Urn:         "is_encrypted",
			DisplayName: "Is Encrypted?",
			Description: "Specify whether this record is encrypted or not.",
			DataType:    "boolean",
			Required:    true,
		},
	},
}

func TestGetAllTagTemplates(t *testing.T) {
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllTagTemplatesRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.GetAllTagTemplatesResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return internal server error if found unexpected error`,
			Request: &compassv1beta1.GetAllTagTemplatesRequest{
				Urn: sampleTemplate.URN,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return(nil, errors.New("unexpected error"))
			},
		},
		{
			Description: `should return ok and templates if found template based on the query`,
			Request: &compassv1beta1.GetAllTagTemplatesRequest{
				Urn: sampleTemplate.URN,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllTagTemplatesResponse) error {
				expected := &compassv1beta1.GetAllTagTemplatesResponse{
					Data: []*compassv1beta1.TagTemplate{
						sampleTemplatePB,
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		}, {
			Description:  `should return all templates if no urn query params`,
			Request:      &compassv1beta1.GetAllTagTemplatesRequest{},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().ReadAll(ctx).Return([]tag.Template{sampleTemplate}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllTagTemplatesResponse) error {
				expected := &compassv1beta1.GetAllTagTemplatesResponse{
					Data: []*compassv1beta1.TagTemplate{
						sampleTemplatePB,
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
			defer mockTemplateRepo.AssertExpectations(t)
			defer mockTagRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService:         service,
				TagTemplateService: templateService,
			})

			got, err := handler.GetAllTagTemplates(ctx, tc.Request)
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

func TestCreateTagTemplate(t *testing.T) {
	validRequest := &compassv1beta1.CreateTagTemplateRequest{
		Urn:         sampleTemplatePB.GetUrn(),
		DisplayName: sampleTemplatePB.GetDisplayName(),
		Description: sampleTemplatePB.GetDescription(),
		Fields:      sampleTemplatePB.GetFields(),
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.CreateTagTemplateResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return already exist if duplicate template`,
			Request:      validRequest,
			ExpectStatus: codes.AlreadyExists,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return(nil, nil)
				ttr.EXPECT().Create(ctx, &sampleTemplate).Return(tag.DuplicateTemplateError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return internal server error if found error during insert`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return(nil, nil)
				ttr.EXPECT().Create(ctx, &sampleTemplate).Return(errors.New("unexpected error during insert"))
			},
		},
		{
			Description:  `should return ok and domain is inserted if found no error`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return(nil, nil)
				ttr.EXPECT().Create(ctx, &sampleTemplate).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.CreateTagTemplateResponse) error {
				expected := &compassv1beta1.CreateTagTemplateResponse{
					Data: sampleTemplatePB,
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
			defer mockTemplateRepo.AssertExpectations(t)
			defer mockTagRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService:         service,
				TagTemplateService: templateService,
			})

			got, err := handler.CreateTagTemplate(ctx, tc.Request)
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

func TestGetTagTemplate(t *testing.T) {
	validRequest := &compassv1beta1.GetTagTemplateRequest{
		TemplateUrn: sampleTemplatePB.GetUrn(),
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.GetTagTemplateResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should returnnot found if template is not found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{}, nil)
			},
		},
		{
			Description:  `should return ok and template if domain template is found`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetTagTemplateResponse) error {
				expected := &compassv1beta1.GetTagTemplateResponse{
					Data: sampleTemplatePB,
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
			defer mockTemplateRepo.AssertExpectations(t)
			defer mockTagRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService:         service,
				TagTemplateService: templateService,
			})

			got, err := handler.GetTagTemplate(ctx, tc.Request)
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

func TestUpdateTagTemplate(t *testing.T) {
	validRequest := &compassv1beta1.UpdateTagTemplateRequest{
		TemplateUrn: sampleTemplatePB.GetUrn(),
		DisplayName: sampleTemplatePB.GetDisplayName(),
		Description: sampleTemplatePB.GetDescription(),
		Fields:      sampleTemplatePB.GetFields(),
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpdateTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
		PostCheck    func(resp *compassv1beta1.UpdateTagTemplateResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return not found if template is not found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{}, nil)
			},
		},
		{
			Description:  `should return invalid argument if there is validation error`,
			Request:      validRequest,
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return(nil, tag.ValidationError{Err: errors.New("validation error")})
			},
		},
		{
			Description:  `should return internal server error if encountered error during update`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return(nil, errors.New("unexpected error"))
			},
		},
		{
			Description:  `should return status ok and its message if successfully updated`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Read(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
				ttr.EXPECT().Update(ctx, sampleTemplate.URN, &sampleTemplate).Run(func(ctx context.Context, templateURN string, template *tag.Template) {
					template.UpdatedAt = time.Now()
				}).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.UpdateTagTemplateResponse) error {
				expectedTemplatePB := sampleTemplatePB
				expectedTemplatePB.UpdatedAt = resp.GetData().GetUpdatedAt()
				expected := &compassv1beta1.UpdateTagTemplateResponse{
					Data: sampleTemplatePB,
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
			defer mockTemplateRepo.AssertExpectations(t)
			defer mockTagRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService:         service,
				TagTemplateService: templateService,
			})

			got, err := handler.UpdateTagTemplate(ctx, tc.Request)
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

func TestDeleteTagTemplate(t *testing.T) {
	validRequest := &compassv1beta1.DeleteTagTemplateRequest{
		TemplateUrn: sampleTemplatePB.GetUrn(),
	}
	type testCase struct {
		Description  string
		Request      *compassv1beta1.DeleteTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagRepository, *mocks.TagTemplateRepository)
	}

	var testCases = []testCase{
		{
			Description:  `should return not found if template is not found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Delete(ctx, sampleTemplate.URN).Return(tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return status ok and template if domain template is found`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, tr *mocks.TagRepository, ttr *mocks.TagTemplateRepository) {
				ttr.EXPECT().Delete(ctx, sampleTemplate.URN).Return(nil)
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
			defer mockTemplateRepo.AssertExpectations(t)
			defer mockTagRepo.AssertExpectations(t)

			templateService := tag.NewTemplateService(mockTemplateRepo)
			service := tag.NewService(mockTagRepo, templateService)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TagService:         service,
				TagTemplateService: templateService,
			})

			_, err := handler.DeleteTagTemplate(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}
