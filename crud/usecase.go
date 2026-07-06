package crud

import (
	"context"

	"github.com/alelmtech/gocore/pagination"
)

type Hooks[E any, ID comparable] struct {
	PreCreate  func(ctx context.Context, entity *E) error
	PostCreate func(ctx context.Context, entity *E) error
	PreUpdate  func(ctx context.Context, entity *E) error
	PreDelete  func(ctx context.Context, id ID) error
}

type BaseUseCase[E any, ID comparable] struct {
	Repo       Repository[E, ID]
	Hooks      Hooks[E, ID]
}

func (uc *BaseUseCase[E, ID]) Create(ctx context.Context, entity *E) error {
	if uc.Hooks.PreCreate != nil {
		if err := uc.Hooks.PreCreate(ctx, entity); err != nil {
			return err
		}
	}
	if err := uc.Repo.Create(ctx, entity); err != nil {
		return err
	}
	if uc.Hooks.PostCreate != nil {
		if err := uc.Hooks.PostCreate(ctx, entity); err != nil {
			return err
		}
	}
	return nil
}

func (uc *BaseUseCase[E, ID]) GetByID(ctx context.Context, id ID) (*E, error) {
	return uc.Repo.FindByID(ctx, id)
}

func (uc *BaseUseCase[E, ID]) Update(ctx context.Context, entity *E) error {
	if uc.Hooks.PreUpdate != nil {
		if err := uc.Hooks.PreUpdate(ctx, entity); err != nil {
			return err
		}
	}
	return uc.Repo.Update(ctx, entity)
}

func (uc *BaseUseCase[E, ID]) Delete(ctx context.Context, id ID) error {
	if uc.Hooks.PreDelete != nil {
		if err := uc.Hooks.PreDelete(ctx, id); err != nil {
			return err
		}
	}
	return uc.Repo.Delete(ctx, id)
}

func (uc *BaseUseCase[E, ID]) List(ctx context.Context, page pagination.Pagination) ([]E, int, error) {
	return uc.Repo.List(ctx, page)
}
