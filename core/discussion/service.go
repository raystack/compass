package discussion

import (
	"context"
	"github.com/raystack/compass/core/namespace"
)

func NewService(discussionRepository Repository) *Service {
	return &Service{
		discussionRepository: discussionRepository,
	}
}

type Service struct {
	discussionRepository Repository
}

func (s *Service) GetDiscussions(ctx context.Context, filter Filter) ([]Discussion, error) {
	return s.discussionRepository.GetAll(ctx, filter)
}
func (s *Service) CreateDiscussion(ctx context.Context, ns *namespace.Namespace, dsc *Discussion) (string, error) {
	return s.discussionRepository.Create(ctx, ns, dsc)
}
func (s *Service) GetDiscussion(ctx context.Context, did string) (Discussion, error) {
	return s.discussionRepository.Get(ctx, did)
}
func (s *Service) PatchDiscussion(ctx context.Context, dsc *Discussion) error {
	return s.discussionRepository.Patch(ctx, dsc)
}
func (s *Service) GetComments(ctx context.Context, discussionID string, filter Filter) ([]Comment, error) {
	return s.discussionRepository.GetAllComments(ctx, discussionID, filter)
}
func (s *Service) CreateComment(ctx context.Context, ns *namespace.Namespace, cmt *Comment) (string, error) {
	return s.discussionRepository.CreateComment(ctx, ns, cmt)
}
func (s *Service) GetComment(ctx context.Context, commentID string, discussionID string) (Comment, error) {
	return s.discussionRepository.GetComment(ctx, commentID, discussionID)
}
func (s *Service) UpdateComment(ctx context.Context, cmt *Comment) error {
	return s.discussionRepository.UpdateComment(ctx, cmt)
}
func (s *Service) DeleteComment(ctx context.Context, commentID string, discussionID string) error {
	return s.discussionRepository.DeleteComment(ctx, commentID, discussionID)
}
