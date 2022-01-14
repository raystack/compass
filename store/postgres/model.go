package postgres

import (
	"time"
)

// Tag is a model for tag value in database table
type Tag struct {
	ID         uint   `gorm:"primaryKey"`
	Value      string `gorm:"not null"`
	RecordType string `gorm:"not null;uniqueIndex:tag_idx_record_urn_record_type_field_id"`
	RecordURN  string `gorm:"not null;uniqueIndex:tag_idx_record_urn_record_type_field_id"`
	FieldID    uint   `gorm:"not null;uniqueIndex:tag_idx_record_urn_record_type_field_id"`
	Field      Field  `gorm:"not null;constraint:OnDelete:CASCADE"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Template is a model for template database table
type Template struct {
	URN         string    `db:"urn"`
	DisplayName string    `db:"display_name"`
	Description string    `db:"description"`
	Fields      []Field   `db:"-"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Field is a model for field tag in database table
type Field struct {
	ID          uint      `db:"id"`
	URN         string    `db:"urn"`
	DisplayName string    `db:"display_name"`
	Description string    `db:"description"`
	DataType    string    `db:"data_type"`
	Options     *string   `db:"options"`
	Required    bool      `db:"required"`
	TemplateURN string    `db:"template_urn"`
	Template    Template  `db:"-"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
