package foosvc

import (
	"context"
	"strings"
	"time"

	"github.com/aatuh/api-toolkit/ports"
)

// Service implements business use-cases for Foo.
type Service struct {
	repo Repo
	tx   ports.TxManager
	log  ports.Logger
	clk  ports.Clock
	ids  ports.IDGen

	timeout  time.Duration
	maxLimit int
}

// ListResult describes a paginated service response.
type ListResult struct {
	Items []Foo
	Total int
}

// New creates a Service with sensible defaults.
func New(
	repo Repo,
	tx ports.TxManager,
	log ports.Logger,
	clk ports.Clock,
	ids ports.IDGen,
) *Service {
	return &Service{
		repo:     repo,
		tx:       tx,
		log:      log,
		clk:      clk,
		ids:      ids,
		timeout:  5 * time.Second,
		maxLimit: 200,
	}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Foo, error) {
	if strings.TrimSpace(in.OrgID) == "" ||
		strings.TrimSpace(in.Namespace) == "" ||
		strings.TrimSpace(in.Name) == "" {
		return nil, ErrInvalid
	}

	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	now := s.clk.Now()
	f := &Foo{
		ID:        s.ids.New(),
		OrgID:     in.OrgID,
		Namespace: in.Namespace,
		Name:      strings.TrimSpace(in.Name),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		return s.repo.Create(txCtx, f)
	})
	if err != nil {
		return nil, s.mapErr(err)
	}
	return f, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Foo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalid
	}
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, s.mapErr(err)
	}
	return f, nil
}

func (s *Service) Update(ctx context.Context, in UpdateInput) (*Foo, error) {
	if strings.TrimSpace(in.ID) == "" || (in.Name == nil) {
		return nil, ErrInvalid
	}
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var out *Foo
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		cur, err := s.repo.GetByID(txCtx, in.ID)
		if err != nil {
			return err
		}
		if in.Name != nil {
			cur.Name = strings.TrimSpace(*in.Name)
		}
		cur.UpdatedAt = s.clk.Now()
		if err := s.repo.Update(txCtx, cur); err != nil {
			return err
		}
		out = cur
		return nil
	})
	if err != nil {
		return nil, s.mapErr(err)
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalid
	}
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	return s.mapErr(s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id)
	}))
}

func (s *Service) List(ctx context.Context, orgID, ns string,
	limit, offset int, search string) (*ListResult, error) {
	if strings.TrimSpace(orgID) == "" || strings.TrimSpace(ns) == "" {
		return nil, ErrInvalid
	}
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	if limit <= 0 || limit > s.maxLimit {
		limit = s.maxLimit
	}
	items, total, err := s.repo.List(ctx, orgID, ns, limit, offset, search)
	if err != nil {
		return nil, s.mapErr(err)
	}
	if total < len(items) {
		total = len(items)
	}
	return &ListResult{Items: items, Total: total}, nil
}

func (s *Service) withTimeout(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}

func (s *Service) mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case err == ErrNotFound:
		return ErrNotFound
	case err == ErrConflict:
		return ErrConflict
	default:
		return ErrInternal
	}
}
