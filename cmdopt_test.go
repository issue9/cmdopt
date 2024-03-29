// SPDX-FileCopyrightText: 2019-2024 caixw
//
// SPDX-License-Identifier: MIT

package cmdopt

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
)

func notFound(s string) string { return "not found " + s }

func TestCmdOpt_New(t *testing.T) {
	a := assert.New(t, false)
	output := new(bytes.Buffer)
	opt := New(output, flag.PanicOnError, "header\noptions\n{{flags}}\ncommands\n{{commands}}\nfooter", nil, notFound)
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

func TestCmdOpt_Exec(t *testing.T) {
	a := assert.New(t, false)

	newOpt := func(a *assert.Assertion) (*CmdOpt, *bytes.Buffer) {
		output := new(bytes.Buffer)
		cmd := func(fs *flag.FlagSet) DoFunc {
			fs.Int("int", 0, "int usage")

			return func(w io.Writer) error {
				_, err := w.Write([]byte("def"))
				return err
			}
		}
		opt := New(output, flag.PanicOnError, "header\noptions\n{{flags}}\ncommands\n{{commands}}\nfooter", cmd, notFound)
		a.NotNil(opt)

		opt.New("test1test1", "test1", "test1 usage\n{{flags}}", func(fs *flag.FlagSet) DoFunc {
			v := false
			fs.BoolVar(&v, "v", false, "usage")

			return func(w io.Writer) error {
				msg := "false"
				if v {
					msg = "true"
				}
				_, err := w.Write([]byte(msg))
				return err
			}
		})

		opt.New("t2", "test2", "test2 usage\nline2", func(fs *flag.FlagSet) DoFunc {
			return func(w io.Writer) error {
				_, err := w.Write([]byte("test2"))
				return err
			}
		})

		cmds := opt.Commands()
		a.Equal([]string{"t2", "test1test1"}, cmds)

		return opt, output
	}

	// Exec test1test1 -v
	opt, output := newOpt(a)
	a.NotError(opt.Exec([]string{"test1test1", "-v", "true"}))
	a.Equal("true", output.String())

	a.PanicString(func() {
		opt.Exec(nil)
	}, "不可多次调用 Exec 方法")

	// Exec test1test1
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{"test1test1"}))
	a.Equal("false", output.String())

	// Exec test1test1 -v true
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{"test1test1", "-v"}))
	a.Equal("true", output.String())

	// Exec t2
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{"t2"}))
	a.Equal("test2", output.String())

	// Exec
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{}))
	a.Equal(output.String(), "def")

	// Exec not-exists
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), notFound("not-exists")))

	// Exec help 未注册
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), notFound("not-exists")))

	// 注册 h
	opt, output = newOpt(a)
	Help(opt, "h", "h-title", "help usage")
	a.NotError(opt.Exec([]string{"h", "test1"})).Equal(output.String(), "not found test1")

	// Exec h not-exists
	opt, output = newOpt(a)
	Help(opt, "h", "h-title", "help usage")
	a.NotError(opt.Exec([]string{"h", "not-exists"}))
	a.True(strings.HasPrefix(output.String(), notFound("not-exists")), output.String())

	// Exec h
	opt, output = newOpt(a)
	Help(opt, "h", "h-title", "help usage")
	a.NotError(opt.Exec([]string{"h"}))
	a.Equal(output.String(), `header
options
  -int int
    	int usage

commands
  h            h-title
  t2           test2
  test1test1   test1

footer
`)

	// Exec h h
	opt, output = newOpt(a)
	Help(opt, "h", "h-title", "help usage")
	a.NotError(opt.Exec([]string{"h", "h"}))
	a.Equal(output.String(), "help usage\n")

	// 非子命令模式 Exec -arg1=xx
	opt, output = newOpt(a)
	a.NotError(opt.Exec([]string{"-int", "5"}))
	a.Equal(output.String(), "def")
}

func TestCmdOpt_Output(t *testing.T) {
	a := assert.New(t, false)

	o1 := new(bytes.Buffer)
	opt := New(o1, flag.ContinueOnError, "", nil, nil)
	opt.New("c1", "title", "usage", func(fs *flag.FlagSet) DoFunc {
		return func(w io.Writer) error {
			_, err := w.Write([]byte("c1"))
			return err
		}
	})
	a.Equal(opt.Output(), o1)

	o2 := new(strings.Builder)
	opt.SetOutput(o2)
	a.Equal(opt.Output(), o2)
	opt.Exec([]string{"c1"})
	a.Equal(o2.String(), "c1").Empty(o1.String())
}
