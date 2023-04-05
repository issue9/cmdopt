// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"io"
	"os"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestCmdOpt_Commands(t *testing.T) {
	a := assert.New(t, false)

	opt := New(os.Stdout, flag.ExitOnError, "usage\nusage", nil, func(s string) string { return "not found " + s })
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
		Equal(usage, "c1 usage")
}
