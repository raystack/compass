package postgres

import "time"

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
	URN         string  `gorm:"primaryKey"`
	DisplayName string  `gorm:"not null"`
	Description string  `gorm:"not null"`
	Fields      []Field `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Field is a model for field tag in database table
type Field struct {
	ID          uint   `gorm:"primaryKey"`
	URN         string `gorm:"not null;uniqueIndex:field_idx_urn_template_urn"`
	DisplayName string `gorm:"not null"`
	Description string `gorm:"not null"`
	DataType    string `gorm:"not null"`
	Options     *string
	Required    bool     `gorm:"not null"`
	TemplateURN string   `gorm:"not null;uniqueIndex:field_idx_urn_template_urn"`
	Template    Template `gorm:"not null;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
