// SPDX-FileCopyrightText: 2019-2026 caixw
//
// SPDX-License-Identifier: MIT

package cmdopt

import (
	"bytes"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/issue9/assert/v5"
)

func TestCmdOpt_New(t *testing.T) {
	a := assert.New(t, false)
	output := new(bytes.Buffer)
	opt := New(&Options{
		Output:        output,
		ErrorHandling: flag.PanicOnError,
		UsageTemplate: "header\noptions\n{{flags}}\ncommands\n{{commands}}\nfooter",
		NotFound:      notFound,
	})
	a.NotNil(opt)

	opt.New("test1test1", "test1", "test1 usage\n{{flags}}", func(fs *flag.FlagSet) DoFunc {
		return func(w io.Writer) error {
			_, err := w.Write([]byte("test1"))
			return err
		}
	})

	a.PanicString(func() {
		opt.New("test1test1", "test1", "usage", func(fs *flag.FlagSet) DoFunc {
			return func(w io.Writer) error { return nil }
		})
	}, "存在相同名称的子命令：test1test1")

	a.PanicString(func() {
		opt.New("t3", "title", "usage", nil)
	}, "参数 cmd 不能为空")

	opt.New("t2", "test2", "test2 usage\nline2", func(fs *flag.FlagSet) DoFunc {
		return func(w io.Writer) error {
			_, err := w.Write([]byte("test2"))
			return err
		}
	})
}

func TestCmdOpt_Commands(t *testing.T) {
	a := assert.New(t, false)

	opt := New(&Options{
		Output:        os.Stdout,
		ErrorHandling: flag.ExitOnError,
		UsageTemplate: "usage\nusage",
		NotFound:      func(s string) string { return "not found " + s },
	})
	a.NotNil(opt)

	a.Length(opt.Commands(), 0)
	_, _, found := opt.Command("cmd")
	a.False(found)

	opt.New("c1", "c1 title", "c1 usage", func(fs *flag.FlagSet) DoFunc { return func(io.Writer) error { return nil } })
	opt.New("c2", "c2 title", "c2 usage", func(fs *flag.FlagSet) DoFunc { return func(io.Writer) error { return nil } })

	a.Length(opt.Commands(), 2)
	_, _, found = opt.Command("cmd")
	a.False(found)
	title, usage, found := opt.Command("c1")
	a.True(found).
		Equal(title, "c1 title").
		Equal(usage, "c1 usage\n")
}
