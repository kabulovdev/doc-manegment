package services

import (
	"context"
	"log/slog"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityService struct {
	repo ports.ActivityLogRepo
}

func NewActivityService(repo ports.ActivityLogRepo) *ActivityService {
	return &ActivityService{repo: repo}
}

type RecordInput struct {
	UserID      primitive.ObjectID
	SubjectType domain.ActivitySubjectType
	SubjectID   primitive.ObjectID
	Action      domain.ActivityAction
	Metadata    map[string]any
	IP          string
	UserAgent   string
}

// Record persists an activity entry. Any error is logged but not returned — the
// caller is a handler that has already served the primary action successfully,
// so activity logging should never fail the request.
func (s *ActivityService) Record(ctx context.Context, in RecordInput) {
	l := &domain.ActivityLog{
		UserID:      in.UserID,
		SubjectType: in.SubjectType,
		SubjectID:   in.SubjectID,
		Action:      in.Action,
		Metadata:    in.Metadata,
		IP:          in.IP,
		UserAgent:   in.UserAgent,
	}
	if err := s.repo.Create(ctx, l); err != nil {
		slog.Warn("activity record failed",
			"err", err,
			"user_id", in.UserID.Hex(),
			"subject", in.SubjectType,
			"action", in.Action,
		)
	}
}

func (s *ActivityService) ListSubject(ctx context.Context, userID, subjectID primitive.ObjectID, limit int64) ([]domain.ActivityLog, error) {
	return s.repo.List(ctx, ports.ActivityFilter{
		UserID:    userID,
		SubjectID: &subjectID,
		Limit:     limit,
	})
}

func (s *ActivityService) ListRecent(ctx context.Context, userID primitive.ObjectID, limit int64) ([]domain.ActivityLog, error) {
	return s.repo.List(ctx, ports.ActivityFilter{
		UserID: userID,
		Limit:  limit,
	})
}
