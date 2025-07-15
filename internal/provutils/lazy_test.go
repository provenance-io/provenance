package provutils

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"cosmossdk.io/log"
)

// LazyTester is a struct that satisfies fmt.Stringer, but marks that it has been .String()ed.
type LazyTester struct {
	Val      string
	Stringed bool
}

func NewLazyTester(val string) *LazyTester {
	return &LazyTester{Val: val}
}

func (l *LazyTester) String() string {
	if l == nil {
		return "<nil>"
	}
	l.Stringed = true
	return l.Val
}

// newInfoLogger returns a new logger with info level that writes to the returned buffer.
func newInfoLogger() (log.Logger, *bytes.Buffer) {
	var buffer bytes.Buffer
	lw := zerolog.ConsoleWriter{
		Out:          &buffer,
		NoColor:      true,
		PartsExclude: []string{"time"}, // Without this, each line starts with "<nil> "
	}
	logger := zerolog.New(lw).Level(zerolog.InfoLevel)
	return log.NewCustomLogger(logger), &buffer
}

func TestLazyTester(t *testing.T) {
	// Since this lazy stuff is dealing with deep-level things that happen automatically,
	// I felt that it was prudent to have a couple tests on the LazyTester. It's used by
	// all the unit tests in this file, so it's good to make sure it's behaving as expected.

	t.Run("marked once String() is called", func(t *testing.T) {
		exp := "I've only got 4 more hours."
		val := NewLazyTester(exp)
		assert.False(t, val.Stringed, "val.Stringed setup check")

		act := val.String()
		assert.Equal(t, exp, act, "result of .String()")
		assert.True(t, val.Stringed, "val.Stringed after .String() is called")
	})

	t.Run("marked once given to sprintf", func(t *testing.T) {
		exp := "Is it naptime yet?"
		val := NewLazyTester(exp)
		assert.False(t, val.Stringed, "val.Stringed setup check")

		act := fmt.Sprintf("%s", val)
		assert.Equal(t, exp, act, "result of fmt.Sprintf")
		assert.True(t, val.Stringed, "val.Stringed after fmt.Sprintf is called")
	})

	t.Run("marked once given to log that is used", func(t *testing.T) {
		str := "There's a hangnail on my pinky."
		val := NewLazyTester(str)
		assert.False(t, val.Stringed, "val.Stringed setup check")

		logger, buffer := newInfoLogger()
		msg := "level just right"
		expOut := "INF " + msg + " val=\"" + str + "\"\n"
		logger.Info(msg, "val", val)
		actOUt := buffer.String()
		assert.True(t, val.Stringed, "val.Stringed after logger.Info")
		assert.Equal(t, expOut, actOUt, "content logged with logger.Info")
	})

	t.Run("not marked when given to log that is not used", func(t *testing.T) {
		str := "My elbow hurts."
		val := NewLazyTester(str)
		assert.False(t, val.Stringed, "val.Stringed setup check")

		logger, buffer := newInfoLogger()
		logger.Debug("level too low", "val", val)
		actOut := buffer.String()
		assert.False(t, val.Stringed, "val.Stringed after logger.Debug")
		assert.Empty(t, actOut, "content logged with logger.Debug")
	})

	t.Run("marked when provided as .With", func(t *testing.T) {
		// This test is more of a demostration of current behavior.
		// It'd be great if .With didn't resolve the value, but it does.
		str := "I'm in the middle of my Wordle."
		val := NewLazyTester(str)
		assert.False(t, val.Stringed, "val.Stringed setup check")

		logger, _ := newInfoLogger()
		logger = logger.With("val", val)
		assert.True(t, val.Stringed, "val.Stringed after logger.With")
		// If the above assertion starts failing, split this sub-test in two, add calls to
		// .Info and .Debug (separate sub-tests), and check the results like the other sub-tests.
	})
}

func TestLazyStringer(t *testing.T) {
	str := "This is really neat!"
	msg := "Hey! Check this out!"
	expOut := "INF " + msg + " lStr=\"" + str + "\"\n"

	val := NewLazyTester(str)
	lStr := NewLazyStringer(val)
	logger, buffer := newInfoLogger()

	// Make sure it is NOT called for Debug level.
	logger.Debug(msg, "lStr", lStr)
	actOut := buffer.String()
	assert.Empty(t, actOut, "debug log output")
	assert.False(t, val.Stringed, "val.Stringed after debug output")
	buffer.Reset()

	// Make sure it IS called for info level.
	logger.Info(msg, "lStr", lStr)
	actOut = buffer.String()
	assert.Equal(t, expOut, actOut, "info log output")
	assert.True(t, val.Stringed, "val.Stringed after info output")
}

func TestLazySprintf(t *testing.T) {
	str1 := "yellow"
	str2 := "purple"
	msg := "Name two colors."
	// No quotes around those two strings because there's no spaces or special chars in them.
	expOut := "INF " + msg + " lStr1=" + str1 + " lStr2=" + str2 + "\n"

	val1 := NewLazyTester(str1)
	val2 := NewLazyTester(str2)
	lStr1 := NewLazyStringer(val1)
	lStr2 := NewLazyStringer(val2)
	logger, buffer := newInfoLogger()

	// Make sure they are NOT called for Debug level.
	logger.Debug(msg, "lStr1", lStr1, "lStr2", lStr2)
	actOut := buffer.String()
	assert.Empty(t, actOut, "debug log output")
	assert.False(t, val1.Stringed, "val1.Stringed after debug output")
	assert.False(t, val2.Stringed, "val2.Stringed after debug output")
	buffer.Reset()

	// Make sure they ARE called for info level.
	logger.Info(msg, "lStr1", lStr1, "lStr2", lStr2)
	actOut = buffer.String()
	assert.Equal(t, expOut, actOut, "info log output")
	assert.True(t, val1.Stringed, "val1.Stringed after info output")
	assert.True(t, val2.Stringed, "val2.Stringed after info output")
}
