package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pnpenv"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/watermill/pnpwatermill"
)

type TestTopicHandler struct {
}

func (t *TestTopicHandler) Name() string {
	return "some_handler"
}
func (t *TestTopicHandler) Topic() string {
	return "test"
}

func (t *TestTopicHandler) Handle(msg *message.Message) error {
	fmt.Println("received", string(msg.Payload))
	return nil
}

func NewTestTopicHandler() *TestTopicHandler {
	return &TestTopicHandler{}
}

func main() {
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("WATERMILL_TRANSPORT", "channel")
	fx.New(
		pnpenv.Module(),
		pnpzap.Module(),
		pnpwatermill.Module(),
		fx.Provide(
			pnpwatermill.HandlerProvider(NewTestTopicHandler),
		),
		fx.Invoke(func(publisher message.Publisher, shutdowner fx.Shutdowner) {
			go func() {
				for i := 0; i < 5; i++ {
					fmt.Println("published", i)
					publisher.Publish("test", &message.Message{Payload: []byte("test " + fmt.Sprint(i))})
					time.Sleep(time.Second)
				}

				go shutdowner.Shutdown()
			}()
		}),
	).Run()
}
