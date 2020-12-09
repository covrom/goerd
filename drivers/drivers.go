package drivers

import (
	"github.com/covrom/goerd/schema"
)

// Driver is the common interface for database drivers
type Driver interface {
	Analyze(*schema.Schema) error
}

// Option is the type for change Config.
type Option func(Driver) error
