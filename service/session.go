package service

import (
	"errors"
	. "github.com/47-11/spotifete/model"
)

type SessionService struct{}

func (s SessionService) GetActiveSessions() ([]*Session, error) {
	return nil, errors.New("not yet implemented")
}

func (s SessionService) GetSessionById(id string) (*Session, error) {
	return nil, errors.New("not yet implemented")
}
