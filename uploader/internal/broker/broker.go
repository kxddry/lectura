package broker

import (
	"context"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Broker interface {
	CheckAlive() error
}

type Writer interface {
	Write(context.Context, uploaded.BrokerRecord) error
	Messages() (chan<- uploaded.BrokerRecord, chan<- error)
	Broker
}
