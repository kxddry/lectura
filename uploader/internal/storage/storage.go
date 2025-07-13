package storage

import (
	"context"
	"github.com/kxddry/lectura/uploader/internal/entities"
)

type S3Storage interface {
	Upload(ctx context.Context, fc entities.File) error
}
