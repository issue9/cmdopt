// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"io"
	"sort"
)

type command struct {
	exec  func(io.Writer, []string) error
	title string
	usage string
}

// Help 注册 help 子命令
func Help(opt *CmdOpt, name, title, usage string) {
	f := func(fs *flag.FlagSet) DoFunc {
		return func(output io.Writer) error {
			if fs.NArg() == 0 {
				_, err := io.WriteString(output, opt.usage())
				return err
			}

			name := fs.Arg(0)
			for _, cmd := range opt.Commands() { // h.opt.Commands() 可以保证顺序一致。
				if cmd == name {
					_, err := io.WriteString(output, opt.commands[cmd].usage)
					return err
				}
			}

			_, err := io.WriteString(output, opt.notFound(name))
			return err
		}
	}

	opt.New(name, title, usage, f)
}

// Commands 返回所有的子命令
func (opt *CmdOpt) Commands() []string {
	keys := make([]string, 0, len(opt.commands))
	for key := range opt.commands {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

// Command 返回指定的命令的说明
func (opt *CmdOpt) Command(name string) (title, usage string, found bool) {
	for k, v := range opt.commands {
		if k == name {
			return v.title, v.usage, true
		}
	}
	return "", "", false
}
