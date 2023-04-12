package internal

import "github.com/google/uuid"

type BoundaryGenerator interface {
	GetBoundary() (string, error)
}

type DefaultBoundaryGenerator struct{}

func (g *DefaultBoundaryGenerator) GetBoundary() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
