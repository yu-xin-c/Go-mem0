package service

import (
	"context"
	"errors"
	"strings"

	"go-mem0/mem0/manager"
	"go-mem0/mem0/model"
	"go-mem0/mem0/planner"
)

type Service struct {
	manager *manager.Manager
	planner planner.Planner
}

func New(manager *manager.Manager, planner planner.Planner) (*Service, error) {
	if manager == nil {
		return nil, errors.New("manager is nil")
	}
	if planner == nil {
		return nil, errors.New("planner is nil")
	}
	return &Service{manager: manager, planner: planner}, nil
}

func (s *Service) UpdateFromMessages(ctx context.Context, userID string, messages []model.Message) ([]model.Memory, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("userID is empty")
	}
	existing, err := s.manager.List(ctx, userID, 100)
	if err != nil {
		return nil, err
	}
	ops, err := s.planner.Plan(ctx, userID, messages, existing)
	if err != nil {
		return nil, err
	}
	if len(ops) == 0 {
		return nil, nil
	}
	out := make([]model.Memory, 0, len(ops))
	for _, op := range ops {
		switch strings.ToLower(op.Kind) {
		case "upsert":
			mem, err := s.manager.Upsert(ctx, userID, op.MemoryID, op.Text, op.Metadata)
			if err != nil {
				return nil, err
			}
			out = append(out, mem)
		case "delete":
			if err := s.manager.Delete(ctx, userID, op.MemoryID); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func (s *Service) Search(ctx context.Context, userID, query string, limit int) ([]model.ScoredMemory, error) {
	return s.manager.Search(ctx, userID, query, limit)
}

func (s *Service) List(ctx context.Context, userID string, limit int) ([]model.Memory, error) {
	return s.manager.List(ctx, userID, limit)
}
