package store

import (
	"database/sql"
	"encoding/json"

	"github.com/raystack/compass/core/principal"
)

type PrincipalModel struct {
	ID          sql.NullString `db:"id"`
	NamespaceID string         `db:"namespace_id"`
	UUID        sql.NullString `db:"uuid"`
	Email       sql.NullString `db:"email"`
	Provider    sql.NullString `db:"provider"`
	Type        sql.NullString `db:"type"`
	Name        sql.NullString `db:"name"`
	Subject     sql.NullString `db:"subject"`
	Metadata    []byte         `db:"metadata"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

func (m *PrincipalModel) toPrincipal() principal.Principal {
	p := principal.Principal{
		ID:        m.ID.String,
		Type:      m.Type.String,
		Name:      m.Name.String,
		Subject:   m.Subject.String,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}
	if len(m.Metadata) > 0 {
		_ = json.Unmarshal(m.Metadata, &p.Metadata)
	}
	// Backward compat: if subject is empty, fall back to uuid
	if p.Subject == "" && m.UUID.Valid {
		p.Subject = m.UUID.String
	}
	if p.Type == "" {
		p.Type = "user"
	}
	return p
}

func newPrincipalModel(p *principal.Principal) PrincipalModel {
	pm := PrincipalModel{}
	if p.ID != "" {
		pm.ID = sql.NullString{String: p.ID, Valid: true}
	}
	if p.Subject != "" {
		pm.Subject = sql.NullString{String: p.Subject, Valid: true}
		// Also set uuid for backward compat with older schema
		pm.UUID = sql.NullString{String: p.Subject, Valid: true}
	}
	if p.Name != "" {
		pm.Name = sql.NullString{String: p.Name, Valid: true}
	}
	if p.Type != "" {
		pm.Type = sql.NullString{String: p.Type, Valid: true}
	} else {
		pm.Type = sql.NullString{String: "user", Valid: true}
	}
	if p.Metadata != nil {
		pm.Metadata, _ = json.Marshal(p.Metadata)
	}
	pm.CreatedAt = sql.NullTime{Time: p.CreatedAt, Valid: true}
	pm.UpdatedAt = sql.NullTime{Time: p.UpdatedAt, Valid: true}

	return pm
}
