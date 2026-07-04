package crud

import (
	"context"

	"github.com/alelmtech/gocore/pagination"
)

type Converter[ENTITY, CREATE_REQ, UPDATE_REQ, RESPONSE any] interface {
	ToEntity(req CREATE_REQ) ENTITY
	UpdateEntity(entity ENTITY, req UPDATE_REQ) ENTITY
	ToResponse(entity ENTITY) RESPONSE
}

type Repository[ENTITY any] interface {
	Create(ctx context.Context, entity *ENTITY) error
	FindByID(ctx context.Context, id string) (*ENTITY, error)
	Update(ctx context.Context, entity *ENTITY) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page pagination.Pagination) ([]ENTITY, int, error)
}

type BaseUseCase[ENTITY, CREATE_REQ, UPDATE_REQ, RESPONSE any] interface {
	Create(ctx context.Context, req CREATE_REQ) (RESPONSE, error)
	GetByID(ctx context.Context, id string) (RESPONSE, error)
	Update(ctx context.Context, id string, req UPDATE_REQ) (RESPONSE, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page pagination.Pagination) ([]RESPONSE, int, error)
}

type BaseUseCaseImpl[ENTITY, CREATE_REQ, UPDATE_REQ, RESPONSE any] struct {
	Repo      Repository[ENTITY]
	Converter Converter[ENTITY, CREATE_REQ, UPDATE_REQ, RESPONSE]
	NewID     func() string
}

func (uc *BaseUseCaseImpl[E, C, U, R]) Create(ctx context.Context, req C) (R, error) {
	entity := uc.Converter.ToEntity(req)
	if err := uc.Repo.Create(ctx, &entity); err != nil {
		return *new(R), err
	}
	return uc.Converter.ToResponse(entity), nil
}

func (uc *BaseUseCaseImpl[E, C, U, R]) GetByID(ctx context.Context, id string) (R, error) {
	entity, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return *new(R), err
	}
	return uc.Converter.ToResponse(*entity), nil
}

func (uc *BaseUseCaseImpl[E, C, U, R]) Update(ctx context.Context, id string, req U) (R, error) {
	entity, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return *new(R), err
	}
	updated := uc.Converter.UpdateEntity(*entity, req)
	if err := uc.Repo.Update(ctx, &updated); err != nil {
		return *new(R), err
	}
	return uc.Converter.ToResponse(updated), nil
}

func (uc *BaseUseCaseImpl[E, C, U, R]) Delete(ctx context.Context, id string) error {
	return uc.Repo.Delete(ctx, id)
}

func (uc *BaseUseCaseImpl[E, C, U, R]) List(ctx context.Context, page pagination.Pagination) ([]R, int, error) {
	entities, total, err := uc.Repo.List(ctx, page)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]R, len(entities))
	for i, e := range entities {
		responses[i] = uc.Converter.ToResponse(e)
	}
	return responses, total, nil
}
