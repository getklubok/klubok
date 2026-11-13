package handler

import (
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func userRegisterParamsToDomain(req *oas.UserRegisterRequest) domain.User {
	return domain.User{
		Email:    string(req.GetEmail()),
		Username: string(req.GetUsername()),
	}
}

func registerResponseFromDomain(u domain.User) *oas.UserRegisterResponse {
	return &oas.UserRegisterResponse{
		ID:       oas.Ulid(u.ID),
		Email:    oas.Email(u.Email),
		Username: oas.Username(u.Username),
	}
}

func meResponseFromDomain(u domain.User) *oas.UserMeResponse {
	return &oas.UserMeResponse{
		ID:       oas.Ulid(u.ID),
		Email:    oas.Email(u.Email),
		Username: oas.Username(u.Username),
		IsFake:   u.IsFake,
	}
}
