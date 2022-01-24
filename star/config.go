package star

const (
	SortKeyCreated             = "created"
	SortKeyUpdated             = "updated"
	SortDirectionKeyAscending  = "asc"
	SortDirectionKeyDescending = "desc"
)

// Config is a config of star domain
type Config struct {
	// Number of relevant results to return
	Size int

	// Offset is a data offset in the table rows
	Offset int

	// Parameter to sort by `CreatedAt` vs `UpdatedAt`
	Sort string

	// SortDirection of sort, ascending/descending
	SortDirection string
}
