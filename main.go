package main

import (
	"github.com/iocgo/sdk"
	"github.com/iocgo/sdk/errors"
	"ppl/wire"
)

func main() {
	ctx := errors.New(nil)
	defer ctx.Throw()
	{
		container := errors.Try1(ctx, func() (c *sdk.Container, err error) {
			c = sdk.NewContainer()
			err = wire.Injects(c)
			return
		})
		if err := container.Run(); err != nil {
			panic(err)
		}
	}
}
