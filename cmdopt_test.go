// SPDX-License-Identifier: MIT

package cmdopt

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
)

func buildDo(text string) DoFunc {
	return func(output io.Writer) error {
		_, err := output.Write([]byte(text))
		return err
	}
}
func notFound(string) string { return "not found" }

func TestCmdOpt(t *testing.T) {
	a := assert.New(t, false)
	output := new(bytes.Buffer)
	opt := New(output, flag.PanicOnError, "header\noptions\n{{flags}}\ncommands\n{{commands}}\nfooter", buildDo("def"), notFound)
	a.NotNil(opt)

	opt.Int("int", 0, "int usage")

	fs1 := opt.New("test1test1", "test1", "test1 usage\n{{flags}}", buildDo("test1"))
	a.NotNil(fs1)
	v := false
	fs1.BoolVar(&v, "v", false, "usage")

	a.Panic(func() {
		opt.New("test1test1", "test1", "usage", buildDo("test1"))
	})

	fs2 := opt.New("t2", "test2", "test2 usage\nline2", buildDo("test2"))
	a.NotNil(fs2)

	cmds := opt.Commands()
	a.Equal([]string{"t2", "test1test1"}, cmds)

	// Exec test1
	a.NotError(opt.Exec([]string{"test1test1", "-v"}))
	a.Equal("test1", output.String())

	// Exec test2
	output.Reset()
	a.NotError(opt.Exec([]string{"t2"}))
	a.Equal("test2", output.String())

	// Exec
	output.Reset()
	a.NotError(opt.Exec([]string{}))
	a.Equal(output.String(), "def")

	// Exec not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), notFound("not-exists")))

	// Exec help 未注册
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), notFound("not-exists")))

	// 注册 h
	opt.Help("h", "h-title", "usage")
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "test1"}))

	// Exec h not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "not-exists"}))
	a.True(strings.HasPrefix(output.String(), notFound("not-exists")))

	// Exec h
	output.Reset()
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
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "h"}))
	a.Equal(output.String(), "usage\n")

	// 非子命令模式 Exec -arg1=xx
	output.Reset()
	a.NotError(opt.Exec([]string{"-int", "5"}))
	a.Equal(output.String(), "def")
}
