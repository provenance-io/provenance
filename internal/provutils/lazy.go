package provutils

import "fmt"

// LazyStringer contains an underlying value (of type T), and is useful when you want to skip calling
// .String() for log statements that are too low-level (e.g. debug when the logger is set to info).
//
// E.g. Use this:
//
//	logger.Debug("something happened", "thing", NewLazyStringer(thing))
//
// Instead of this:
//
//	logger.Debug("something happened", "thing", thing.String())
//
// In the former, thing.String() is only called if the log statement is generated.
// In the latter, thing.String() is called every time, regardless of log-level.
//
// Usually, you'd just leave off the .String() provide on the arg. Sometimes, though, the value gets expanded
// using %v and you'd rather have the .String() version in the log statement (e.g. sdk.Coins).
//
// This doesn't do anything when added to a logger using .With(...) because .With resolves it immediately anyway.
//
// See also: LazySprintf.
type LazyStringer[T fmt.Stringer] struct {
	val T
}

// NewLazyStringer defers calling Val.String() until the Stringer interface is invoked.
// This is particularly useful for avoiding calling .String() when debugging is not active.
//
// Do not use this to provide values to a logger using logger.With(...) because it'll be resolved
// immediately even if the logger doesn't output anything, so it's best to just provide what you want directly.
//
// See also: NewLazySprintf.
func NewLazyStringer[T fmt.Stringer](val T) *LazyStringer[T] {
	return &LazyStringer[T]{val: val}
}

func (l *LazyStringer[T]) String() string {
	return l.val.String()
}

// LazySprintf contains the format and args of a desired call to Sprintf which is only invoked when
// .String() is called. It's useful when you want to skip calling Sprintf for log statements that
// are too low-level (e.g. debug when the logger is set to info).
//
// E.g. Use this:
//
//	logger.Debug("something happened", "progress", NewLazySprintf("[%d/%d]", i+1, total))
//
// Instead of this:
//
//	logger.Debug("something happened", "progress", fmt.Sprintf("[%d/%d]", i+1, total))
//
// In the former, the fmt.Sprintf call is only made if the log statement is generated.
// In the latter, fmt.Sprintf is called every time, regardless of log-level.
//
// This doesn't do anything when added to a logger using .With(...) because .With resolves it immediately anyway.
//
// See also: LazyStringer
type LazySprintf struct {
	format string
	args   []interface{}
}

// NewLazySprintf defers fmt.Sprintf until the Stringer interface is invoked.
// This is particularly useful for avoiding calling Sprintf when debugging is not active.
//
// Do not use this to provide values to a logger using logger.With(...) because it'll be resolved
// immediately even if the logger doesn't output anything, so it's best to just provide what you want directly.
//
// See also: NewLazyStringer.
func NewLazySprintf(format string, args ...interface{}) *LazySprintf {
	return &LazySprintf{format, args}
}

func (l *LazySprintf) String() string {
	return fmt.Sprintf(l.format, l.args...)
}
