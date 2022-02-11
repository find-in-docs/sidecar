package connection

import "context"

type Effector func(context.Context) (*Message, error)
