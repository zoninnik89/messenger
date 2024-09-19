package main

import "context"

type PubSubService interface {
	Subscribe(context.Context, chan string)
	Publish(context.Context, string)
}
