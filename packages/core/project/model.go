package project

import (
	"time"

	devcontext "devctx/packages/core/context"
)

// Path identifies a project by its canonical filesystem path.
//
// This model assumes callers have already normalized the path. It does not
// access the filesystem or decide whether a path is valid.
type Path string

// Binding represents a persistent association between a project and context.
type Binding struct {
	ProjectPath Path
	ContextID   devcontext.ID
	CreatedAt   time.Time
}
