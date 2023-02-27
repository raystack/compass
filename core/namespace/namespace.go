package namespace

//go:generate mockery --name=StorageRepository -r --case underscore --with-expecter --structname NamespaceStorageRepository --filename storage_repository.go --output=./mocks
//go:generate mockery --name=DiscoveryRepository -r --case underscore --with-expecter --structname NamespaceDiscoveryRepository --filename discovery_repository.go --output=./mocks
import (
	"context"
	"github.com/google/uuid"
)

var (
	// DefaultNamespace is used for compass single tenant applications
	DefaultNamespace = &Namespace{
		ID:       uuid.Nil,
		Name:     "default",
		State:    SharedState,
		Metadata: map[string]interface{}{},
	}
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	// PendingState could be used for tenants which are not ready for use at the moment
	PendingState State = "pending"
	// SharedState is used for default small scale tenants
	SharedState State = "shared"
	// DedicatedState is for large tenants
	DedicatedState State = "dedicated"
	// UpgradeState is used when a shared tenant is getting upgraded to dedicated
	// TODO: *impt* : support of upgrading a shared to dedicated tenant is not implemented yet
	UpgradeState State = "upgrade"
)

type Namespace struct {
	ID uuid.UUID `json:"id"`
	// Name should be at least couple of letters ideally as it will be unique same as ID
	Name string `json:"name"`

	State    State                  `json:"state"`
	Metadata map[string]interface{} `json:"metadata"`
}

type StorageRepository interface {
	Create(context.Context, *Namespace) (string, error)
	Update(context.Context, *Namespace) error
	GetByID(context.Context, uuid.UUID) (*Namespace, error)
	GetByName(context.Context, string) (*Namespace, error)
	List(context.Context) ([]*Namespace, error)
}

type DiscoveryRepository interface {
	CreateNamespace(context.Context, *Namespace) error
}
