// SPDX-License-Identifier: MIT

package cmdopt

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

func buildDo(text string) DoFunc {
	return func(output io.Writer) error {
		_, err := output.Write([]byte(text))
		return err
	}
}

func usage(output io.Writer) error {
	_, err := output.Write([]byte("usage"))
	return err
}

func TestOptCmd(t *testing.T) {
	a := assert.New(t)
	output := new(bytes.Buffer)
	opt := New(output, flag.PanicOnError, usage, func(string) string { return "not found" })
	a.NotNil(opt)

	fs1 := opt.New("test1", buildDo("test1"), nil)
	a.NotNil(fs1)
	v := false
	fs1.BoolVar(&v, "v", false, "usage")

	a.Panic(func() {
		opt.New("test1", buildDo("test1"), nil)
	})

	fs2 := opt.New("test2", buildDo("test2"), nil)
	a.NotNil(fs2)

	cmds := opt.Commands()
	a.Equal([]string{"test1", "test2"}, cmds)

	// Exec test1
	a.NotError(opt.Exec([]string{"test1", "-v"}))
	a.Equal("test1", output.String())

	// Exec test2
	output.Reset()
	a.NotError(opt.Exec([]string{"test2"}))
	a.Equal("test2", output.String())

	// Exec
	output.Reset()
	a.NotError(opt.Exec([]string{}))
	a.Equal("usage", output.String())

	// Exec not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), string(opt.notFound("not-exists"))))

	// Exec help 未注册
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), string(opt.notFound("not-exists"))))

	// 注册 h
	opt.Help("h", func(w io.Writer) error {
		_, err := w.Write([]byte("usage"))
		return err
	})
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "test1"}))

	// Exec h not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "not-exists"}))
	a.True(strings.HasPrefix(output.String(), string(opt.notFound("not-exists"))))

	// Exec h
	output.Reset()
	a.NotError(opt.Exec([]string{"h", ""}))
	a.True(strings.HasPrefix(output.String(), string(opt.notFound(""))))

	// Exec h h
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "h"}))
	a.Equal(output.String(), "usage")
}
