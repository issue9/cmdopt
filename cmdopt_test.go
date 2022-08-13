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

func newCmdOpt(
	output io.Writer,
	errHandling flag.ErrorHandling,
	header, footer, options, commands string,
	notFound func(string) string,
) *CmdOpt {
	return &CmdOpt{
		ErrorHandling: errHandling,
		Output:        output,
		Header:        header,
		Footer:        footer,
		OptionsTitle:  options,
		CommandsTitle: commands,
		NotFound:      notFound,
	}
}

func buildDo(text string) DoFunc {
	return func(output io.Writer) error {
		_, err := output.Write([]byte(text))
		return err
	}
}

func TestCmdOpt(t *testing.T) {
	a := assert.New(t, false)
	output := new(bytes.Buffer)
	opt := newCmdOpt(output, flag.PanicOnError, "header\n", "footer\n", "options", "commands", func(string) string { return "not found" })
	a.NotNil(opt)

	fs1 := opt.New("test1test1", "test1 usage", buildDo("test1"))
	a.NotNil(fs1)
	v := false
	fs1.BoolVar(&v, "v", false, "usage")

	a.Panic(func() {
		opt.New("test1test1", "usage", buildDo("test1"))
	})

	fs2 := opt.New("t2", "test2 usage\nline2", buildDo("test2"))
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
	a.Equal(output.String(), `header

commands
    t2           test2 usage
    test1test1   test1 usage

footer
`)

	// Exec not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), opt.NotFound("not-exists")))

	// Exec help 未注册
	output.Reset()
	a.NotError(opt.Exec([]string{"not-exists"}))
	a.True(strings.HasPrefix(output.String(), opt.NotFound("not-exists")))

	// 注册 h
	opt.Help("h", "usage")
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "test1"}))

	// Exec h not-exists
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "not-exists"}))
	a.True(strings.HasPrefix(output.String(), opt.NotFound("not-exists")))

	// Exec h
	output.Reset()
	a.NotError(opt.Exec([]string{"h", ""}))
	a.True(strings.HasPrefix(output.String(), opt.NotFound("")))

	// Exec h h
	output.Reset()
	a.NotError(opt.Exec([]string{"h", "h"}))
	a.Equal(output.String(), "usage")
}
