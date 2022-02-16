package conn

import "context"

type Circuit func(context.Context, *Message) (*Message, error)
