package service

import (
	"errors"

	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
)

var (
	ErrUserExists            = errors.New("USER EXISTS")
	ErrUserCredentialInvalid = errors.New("USER CREDENTIAL INVALID")
)

type Service struct {
}

func (s *Service) GetPesonByCredential(login string, pass string) (models.Person, error) {

	return models.Person{}, nil
}

func (s *Service) CreatePeson(p models.Person) (models.Person, error) {

	return models.Person{}, nil
}

func NewService(log logger.Lg) (Service, error) {

	return Service{}, nil

}
