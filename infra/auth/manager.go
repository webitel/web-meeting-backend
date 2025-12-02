package auth

import (
	"context"
)

type Manager interface {
	AuthorizeFromContext(ctx context.Context, mainObjClassName string, mainAccessMode AccessMode) (Session, error)
}
