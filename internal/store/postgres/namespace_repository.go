package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/odpf/compass/core/namespace"
)

const (
	namespaceTable = "namespaces"
)

var (
	ErrNamespaceNotFound = errors.New("namespace not found")
)

type NamespaceRepository struct {
	client *Client
}

// Create insert a new namespace in the database
func (n *NamespaceRepository) Create(ctx context.Context, ns *namespace.Namespace) (string, error) {
	var nsID string
	if ns == nil || ns.Name == "" {
		return "", errors.New("invalid namespace")
	}
	nsModel, err := BuildNamespaceModel(*ns)
	if err != nil {
		return "", err
	}

	query, args, err := sq.Insert(namespaceTable).Columns("id", "name", "state", "metadata").
		Values(nsModel.ID, nsModel.Name, nsModel.State, nsModel.Metadata).Suffix("RETURNING \"id\"").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return "", err
	}

	err = n.client.QueryFn(ctx, func(conn *sqlx.Conn) error {
		return conn.QueryRowxContext(ctx, query, args...).Scan(&nsID)
	})
	if err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", errors.New("namespace already exists")
		}
		return "", err
	}
	if nsID == "" {
		return "", fmt.Errorf("error Namespace ID is empty from DB")
	}
	return nsID, nil
}

// Update an existing namespace to the database
func (n *NamespaceRepository) Update(ctx context.Context, ns *namespace.Namespace) error {
	nsModel, err := BuildNamespaceModel(*ns)
	if err != nil {
		return err
	}

	query, args, err := sq.Update(namespaceTable).Set("state", nsModel.State).
		Set("metadata", nsModel.Metadata).Where("id = ?", nsModel.ID).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	_, err = n.client.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error running update namespace query: %w", err)
	}
	return nil
}

func (n *NamespaceRepository) GetByName(ctx context.Context, name string) (*namespace.Namespace, error) {
	query, args, err := sq.Select("*").From(namespaceTable).Where("name = ?", name).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	var nsModel NamespaceModel
	if err := n.client.GetContext(ctx, &nsModel, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNamespaceNotFound
		}
		return nil, err
	}
	return nsModel.toNamespace()
}

// GetByID retrieves namespace given the uuid
func (n *NamespaceRepository) GetByID(ctx context.Context, uuid uuid.UUID) (*namespace.Namespace, error) {
	query, args, err := sq.Select("*").From(namespaceTable).Where("id = ?", uuid).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	var nsModel NamespaceModel
	if err := n.client.GetContext(ctx, &nsModel, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNamespaceNotFound
		}
		return nil, err
	}
	return nsModel.toNamespace()
}

// List retrieves all namespaces
func (n *NamespaceRepository) List(ctx context.Context) ([]*namespace.Namespace, error) {
	var nsModels []*NamespaceModel
	query, _, err := sq.Select(`*`).From(namespaceTable).ToSql()
	if err != nil {
		return nil, err
	}
	if err := n.client.SelectContext(ctx, &nsModels, query); err != nil {
		return nil, err
	}
	var namespaces []*namespace.Namespace
	for _, nsModel := range nsModels {
		ns, err := nsModel.toNamespace()
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

// NewNamespaceRepository initializes namespace repository
func NewNamespaceRepository(c *Client) *NamespaceRepository {
	return &NamespaceRepository{
		client: c,
	}
}
