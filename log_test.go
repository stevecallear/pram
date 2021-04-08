package pram_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/stevecallear/pram"
)

func TestLogf(t *testing.T) {
	t.Run("should not panic if a nil logger is used", func(t *testing.T) {
		pram.SetLogger(nil)
		pram.Logf("value")
	})

	t.Run("should use the specified logger", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		pram.SetLogger(log.New(buf, "", 0))
		defer pram.SetLogger(nil)

		pram.Logf("value: %s", "expected")

		if act, exp := buf.String(), "value: expected\n"; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}
	})
}
