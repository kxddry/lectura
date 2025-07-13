package broker

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Writer interface {
	Write(context.Context, uploaded.KafkaRecord) error
}
