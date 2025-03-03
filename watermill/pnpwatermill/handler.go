package pnpwatermill

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/fxutil"
)

// HandlerProvider wraps Handler constructor so it can be used by pnpwatermill module
func HandlerProvider(target any) any {
	return fxutil.GroupProvider[Handler]("pnpwatermill.handlers", target)
}

type Handler interface {
	Name() string
	Topic() string
	Handle(msg *message.Message) error
}
