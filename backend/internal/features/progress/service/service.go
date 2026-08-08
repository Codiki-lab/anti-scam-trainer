package service

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
