package broker

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
)

type Broker interface {
	CheckAlive() error
}

type Reader interface {
	Read(context.Context) (transcribed.BrokerRecord, error)
}

type Writer interface {
	Write(context.Context, summarized.BrokerRecord) error
	Broker
}
