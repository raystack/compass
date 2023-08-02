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
	"google.golang.org/protobuf/types/known/timestamppb"
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
			Description: "The classification of this asset",
			DataType:    "enumerated",
			Required:    true,
			Options:     []string{"Public", "Restricted"},
		},
		{
			ID:          2,
			URN:         "is_encrypted",
			DisplayName: "Is Encrypted?",
			Description: "Specify whether this asset is encrypted or not.",
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
			Description: "The classification of this asset",
			DataType:    "enumerated",
			Required:    true,
			Options:     []string{"Public", "Restricted"},
		},
		{
			Id:          2,
			Urn:         "is_encrypted",
			DisplayName: "Is Encrypted?",
			Description: "Specify whether this asset is encrypted or not.",
			DataType:    "boolean",
			Required:    true,
		},
	},
}

func TestGetAllTagTemplates(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllTagTemplatesRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.GetAllTagTemplatesResponse) error
	}

	testCases := []testCase{
		{
			Description: `should return internal server error if found unexpected error`,
			Request: &compassv1beta1.GetAllTagTemplatesRequest{
				Urn: sampleTemplate.URN,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().GetTemplates(ctx, sampleTemplate.URN).Return(nil, errors.New("unexpected error"))
			},
		},
		{
			Description: `should return ok and templates if found template based on the query`,
			Request: &compassv1beta1.GetAllTagTemplatesRequest{
				Urn: sampleTemplate.URN,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().GetTemplates(ctx, sampleTemplate.URN).Return([]tag.Template{sampleTemplate}, nil)
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
		{
			Description:  `should return all templates if no urn query params`,
			Request:      &compassv1beta1.GetAllTagTemplatesRequest{},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().GetTemplates(ctx, "").Return([]tag.Template{sampleTemplate}, nil)
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

			handler := NewAPIServer(APIServerDeps{
				TagSvc:         mockTagSvc,
				TagTemplateSvc: mockTemplateSvc,
				UserSvc:        mockUserSvc,
				Logger:         logger,
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
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.CreateTagTemplateRequest{
			Urn:         sampleTemplatePB.GetUrn(),
			DisplayName: sampleTemplatePB.GetDisplayName(),
			Description: sampleTemplatePB.GetDescription(),
			Fields:      sampleTemplatePB.GetFields(),
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.CreateTagTemplateResponse) error
	}

	testCases := []testCase{
		{
			Description: `should return invalid argument if urn is empty`,
			Request: &compassv1beta1.CreateTagTemplateRequest{
				Urn:         "",
				DisplayName: sampleTemplatePB.GetDisplayName(),
				Description: sampleTemplatePB.GetDescription(),
				Fields:      sampleTemplatePB.GetFields(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if display name is empty`,
			Request: &compassv1beta1.CreateTagTemplateRequest{
				Urn:         sampleTemplatePB.GetUrn(),
				DisplayName: "",
				Description: sampleTemplatePB.GetDescription(),
				Fields:      sampleTemplatePB.GetFields(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if description is empty`,
			Request: &compassv1beta1.CreateTagTemplateRequest{
				Urn:         sampleTemplatePB.GetUrn(),
				DisplayName: sampleTemplatePB.GetDisplayName(),
				Description: "",
				Fields:      sampleTemplatePB.GetFields(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if fields is nil`,
			Request: &compassv1beta1.CreateTagTemplateRequest{
				Urn:         sampleTemplatePB.GetUrn(),
				DisplayName: sampleTemplatePB.GetDisplayName(),
				Description: sampleTemplatePB.GetDescription(),
				Fields:      nil,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return already exist if duplicate template`,
			Request:      validRequest,
			ExpectStatus: codes.AlreadyExists,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().CreateTemplate(ctx, &sampleTemplate).Return(tag.DuplicateTemplateError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return internal server error if found error during insert`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().CreateTemplate(ctx, &sampleTemplate).Return(errors.New("unexpected error during insert"))
			},
		},
		{
			Description:  `should return ok and domain is inserted if found no error`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().CreateTemplate(ctx, &sampleTemplate).Return(nil)
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

			handler := NewAPIServer(APIServerDeps{
				TagSvc:         mockTagSvc,
				TagTemplateSvc: mockTemplateSvc,
				UserSvc:        mockUserSvc,
				Logger:         logger,
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
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.GetTagTemplateRequest{
			TemplateUrn: sampleTemplatePB.GetUrn(),
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.GetTagTemplateResponse) error
	}

	testCases := []testCase{
		{
			Description:  `should return not found if template is not found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().GetTemplate(ctx, sampleTemplate.URN).Return(tag.Template{}, tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return internal error if there is error in fetching the template`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().GetTemplate(ctx, sampleTemplate.URN).Return(tag.Template{}, errors.New("some error"))
			},
		},
		{
			Description:  `should return ok and template if domain template is found`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().GetTemplate(ctx, sampleTemplate.URN).Return(sampleTemplate, nil)
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

			handler := NewAPIServer(APIServerDeps{
				TagSvc:         mockTagSvc,
				TagTemplateSvc: mockTemplateSvc,
				UserSvc:        mockUserSvc,
				Logger:         logger,
			})

			got, err := handler.GetTagTemplate(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
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
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.UpdateTagTemplateRequest{
			TemplateUrn: sampleTemplatePB.GetUrn(),
			DisplayName: sampleTemplatePB.GetDisplayName(),
			Description: sampleTemplatePB.GetDescription(),
			Fields:      sampleTemplatePB.GetFields(),
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpdateTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
		PostCheck    func(resp *compassv1beta1.UpdateTagTemplateResponse) error
	}

	testCases := []testCase{
		{
			Description:  `should return not found if template is not found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().UpdateTemplate(ctx, sampleTemplate.URN, &sampleTemplate).Return(tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description: `should return invalid argument if display name is empty`,
			Request: &compassv1beta1.UpdateTagTemplateRequest{
				TemplateUrn: sampleTemplatePB.GetUrn(),
				DisplayName: "",
				Description: sampleTemplatePB.GetDescription(),
				Fields:      sampleTemplatePB.GetFields(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if description is empty`,
			Request: &compassv1beta1.UpdateTagTemplateRequest{
				TemplateUrn: sampleTemplatePB.GetUrn(),
				DisplayName: sampleTemplatePB.GetDisplayName(),
				Description: "",
				Fields:      sampleTemplatePB.GetFields(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if fields is empty`,
			Request: &compassv1beta1.UpdateTagTemplateRequest{
				TemplateUrn: sampleTemplatePB.GetUrn(),
				DisplayName: sampleTemplatePB.GetDisplayName(),
				Description: sampleTemplatePB.GetDescription(),
				Fields:      nil,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  `should return invalid argument if there is validation error`,
			Request:      validRequest,
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().UpdateTemplate(ctx, sampleTemplate.URN, &sampleTemplate).Return(tag.ValidationError{Err: errors.New("validation error")})
			},
		},
		{
			Description:  `should return internal server error if encountered error during update`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().UpdateTemplate(ctx, sampleTemplate.URN, &sampleTemplate).Return(errors.New("unexpected error"))
			},
		},
		{
			Description:  `should return status ok and its message if successfully updated`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().UpdateTemplate(ctx, sampleTemplate.URN, &sampleTemplate).Run(func(ctx context.Context, templateURN string, template *tag.Template) {
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

			handler := NewAPIServer(APIServerDeps{
				TagSvc:         mockTagSvc,
				TagTemplateSvc: mockTemplateSvc,
				UserSvc:        mockUserSvc,
				Logger:         logger,
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
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		validRequest = &compassv1beta1.DeleteTagTemplateRequest{
			TemplateUrn: sampleTemplatePB.GetUrn(),
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.DeleteTagTemplateRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.TagService, *mocks.TagTemplateService)
	}

	testCases := []testCase{
		{
			Description:  `should return not found if template is not found`,
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().DeleteTemplate(ctx, sampleTemplate.URN).Return(tag.TemplateNotFoundError{URN: sampleTemplate.URN})
			},
		},
		{
			Description:  `should return internal error if there is an error in deleting the template`,
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().DeleteTemplate(ctx, sampleTemplate.URN).Return(errors.New("internal error"))
			},
		},
		{
			Description:  `should return status ok and template if domain template is found`,
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ts *mocks.TagService, tts *mocks.TagTemplateService) {
				tts.EXPECT().DeleteTemplate(ctx, sampleTemplate.URN).Return(nil)
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

			handler := NewAPIServer(APIServerDeps{
				TagSvc:         mockTagSvc,
				TagTemplateSvc: mockTemplateSvc,
				UserSvc:        mockUserSvc,
				Logger:         logger,
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

func TestTemplateToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Template    tag.Template
		ExpectProto *compassv1beta1.TagTemplate
	}

	testCases := []testCase{
		{
			Title:       "should return no timestamp pb and empty template field pb if timestamp and template field are empty",
			Template:    tag.Template{URN: "urn", DisplayName: "display-name", Description: "description"},
			ExpectProto: &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description"},
		},
		{
			Title:       "should return timestamp pb and template field pb if timestamp and template field are not empty",
			Template:    tag.Template{URN: "urn", DisplayName: "display-name", Description: "description", Fields: []tag.Field{{ID: 12, URN: "urn1"}}, CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description", Fields: []*compassv1beta1.TagTemplateField{{Id: 12, Urn: "urn1"}}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := tagTemplateToProto(tc.Template)
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestTagTemplateFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.TagTemplate
		Expect tag.Template
	}

	testCases := []testCase{
		{
			Title:  "should return non empty time.Time and field if timestamp pb and field pb are not empty or zero",
			PB:     &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description", Fields: []*compassv1beta1.TagTemplateField{{Id: 12, Urn: "urn1"}}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: tag.Template{URN: "urn", DisplayName: "display-name", Description: "description", Fields: []tag.Field{{ID: 12, URN: "urn1"}}, CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time and empty field if timestamp pb and field pb are empty or zero",
			PB:     &compassv1beta1.TagTemplate{Urn: "urn", DisplayName: "display-name", Description: "description"},
			Expect: tag.Template{URN: "urn", DisplayName: "display-name", Description: "description"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := tagTemplateFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}

func TestTemplateFieldToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Field       tag.Field
		ExpectProto *compassv1beta1.TagTemplateField
	}

	testCases := []testCase{
		{
			Title:       "should return no timestamp pb if timestamp is empty",
			Field:       tag.Field{ID: 123, URN: "urn"},
			ExpectProto: &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn"},
		},
		{
			Title:       "should return timestamp pb if timestamp is not empty or zero",
			Field:       tag.Field{ID: 123, URN: "urn", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := tagTemplateFieldToProto(tc.Field)
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestTagTemplateFieldFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.TagTemplateField
		Expect tag.Field
	}

	testCases := []testCase{
		{
			Title:  "should return non empty time.Time if timestamp pb is not empty",
			PB:     &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: tag.Field{ID: 123, URN: "urn", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time if timestamp pb is empty or zero",
			PB:     &compassv1beta1.TagTemplateField{Id: 123, Urn: "urn"},
			Expect: tag.Field{ID: 123, URN: "urn"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := tagTemplateFieldFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
