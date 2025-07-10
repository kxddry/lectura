package broker

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Broker interface {
	CheckAlive() error
}

type Reader interface {
	Read(context.Context) (uploaded.BrokerRecord, error)
}

type Writer interface {
	Write(context.Context, uploaded.BrokerRecord) error
	Broker
}
