package postgres

import (
	"github.com/google/uuid"
	"github.com/odpf/compass/core/namespace"
	"time"
)

type NamespaceModel struct {
	ID        uuid.UUID  `db:"id"`
	Name      string     `db:"name"`
	State     string     `db:"state"`
	Metadata  JSONMap    `db:"metadata"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func BuildNamespaceModel(ns namespace.Namespace) (*NamespaceModel, error) {
	return &NamespaceModel{
		ID:       ns.ID,
		Name:     ns.Name,
		State:    ns.State.String(),
		Metadata: ns.Metadata,
	}, nil
}

func (a NamespaceModel) toNamespace() (*namespace.Namespace, error) {
	return &namespace.Namespace{
		ID:       a.ID,
		Name:     a.Name,
		State:    namespace.State(a.State),
		Metadata: a.Metadata,
	}, nil
}
