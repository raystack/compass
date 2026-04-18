package principal_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/principal"
	"github.com/raystack/compass/core/principal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestValidatePrincipal(t *testing.T) {
	ns := &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "tenant",
		State:    namespace.SharedState,
		Metadata: nil,
	}
	type testCase struct {
		Description string
		Subject     string
		Name        string
		Type        string
		Setup       func(ctx context.Context, subject string, repo *mocks.PrincipalRepository)
		ExpectErr   error
	}

	var testCases = []testCase{
		{
			Description: "should return no principal error when subject is empty",
			Subject:     "",
			ExpectErr:   principal.ErrNoPrincipalInformation,
		},
		{
			Description: "should return principal ID when successfully found from DB",
			Subject:     "a-subject",
			Setup: func(ctx context.Context, subject string, repo *mocks.PrincipalRepository) {
				repo.EXPECT().GetBySubject(ctx, subject).Return(principal.Principal{ID: "principal-id", Subject: subject}, nil)
			},
			ExpectErr: nil,
		},
		{
			Description: "should return error if GetBySubject returns nil error and empty ID",
			Subject:     "a-subject",
			Setup: func(ctx context.Context, subject string, repo *mocks.PrincipalRepository) {
				repo.EXPECT().GetBySubject(ctx, subject).Return(principal.Principal{}, nil)
			},
			ExpectErr: errors.New("fetched principal ID from DB is empty"),
		},
		{
			Description: "should return principal ID when not found from DB but can create new one",
			Subject:     "new-subject",
			Setup: func(ctx context.Context, subject string, repo *mocks.PrincipalRepository) {
				repo.EXPECT().GetBySubject(ctx, subject).Return(principal.Principal{}, principal.NotFoundError{Subject: subject})
				repo.EXPECT().Upsert(ctx, ns, &principal.Principal{Subject: subject, Type: "user"}).Return("some-id", nil)
			},
			ExpectErr: nil,
		},
		{
			Description: "should return error when not found from DB and can't create new one",
			Subject:     "error-subject",
			Setup: func(ctx context.Context, subject string, repo *mocks.PrincipalRepository) {
				mockErr := errors.New("error upserting principal")
				repo.EXPECT().GetBySubject(ctx, subject).Return(principal.Principal{}, mockErr)
			},
			ExpectErr: errors.New("error upserting principal"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(mocks.PrincipalRepository)

			if tc.Setup != nil {
				tc.Setup(ctx, tc.Subject, mockRepo)
			}

			svc := principal.NewService(mockRepo)

			_, err := svc.ValidatePrincipal(ctx, ns, tc.Subject, tc.Name, tc.Type)

			assert.Equal(t, tc.ExpectErr, err)
		})
	}
}
