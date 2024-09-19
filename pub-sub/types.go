package main

import "context"

type PubSubServiceInterface interface {
	Subscribe(context.Context, chan string)
	Publish(context.Context, string)
}
