package conn

import "context"

type Effector func(context.Context) (*Message, error)
