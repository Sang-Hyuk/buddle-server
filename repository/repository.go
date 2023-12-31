package repository

import "github.com/pkg/errors"

type Repository interface {
	Product() ProductRepository
	User() UserRepository
	AfterService() AfterServiceRepository
}

type repository struct {
	product      ProductRepository
	user         UserRepository
	afterService AfterServiceRepository
}

func (r repository) Product() ProductRepository {
	return r.product
}

func (r repository) User() UserRepository {
	return r.user
}

func (r repository) AfterService() AfterServiceRepository {
	return r.afterService
}

func (r repository) Validate() error {
	switch {
	case r.Product() == nil:
		return errors.New("product repository is nil")
	case r.User() == nil:
		return errors.New("user repository is nil")
	case r.AfterService() == nil:
		return errors.New("product repository is nil")
	}

	return nil
}

func NewRepository() (Repository, error) {
	r := &repository{
		product: NewProductRepository(),
		user:    NewUserRepository(),
		afterService: NewAfterServiceRepository(),
	}

	if err := r.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to create repository")
	}

	return r, nil
}
