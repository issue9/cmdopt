// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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

func usage(output io.Writer) {
	output.Write([]byte("usage"))
}

func TestOptCmd(t *testing.T) {
	a := assert.New(t)
	output := new(bytes.Buffer)
	opt := New(output, flag.ExitOnError, usage)
	a.NotNil(opt)

	fs1 := opt.New("test1", buildDo("test1"))
	a.NotNil(fs1)

	a.Panic(func() {
		opt.New("test1", buildDo("test1"))
	})

	fs2 := opt.New("test2", buildDo("test2"))
	a.NotNil(fs2)

	cmds := opt.Commands()
	a.Equal([]string{"test1", "test2"}, cmds)

	// Exec test1
	a.NotError(opt.Exec([]string{"test1", "xxx"}))
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
	a.True(strings.HasPrefix(output.String(), string(notFound)))

	// Exec help 未注册
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), string(notFound)))

	// 注册 Help
	opt.Help("h")
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "test1"}))

	// Exec Help not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "not-exists"}))
	a.True(strings.HasPrefix(output.String(), string(notFound)))

	// Exec Help
	output.Reset()
	a.NotError(opt.Exec([]string{"h", ""}))
	a.True(strings.HasPrefix(output.String(), string(notFound)))
}
