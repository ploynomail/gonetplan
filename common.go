package gonetplan

import "context"

type Logger interface {
	Debugf(ctx context.Context, format string, v ...interface{})
	Errorf(ctx context.Context, format string, v ...interface{})
}
