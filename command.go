// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"io"
	"sort"
)

type command struct {
	fs    *flag.FlagSet
	do    DoFunc
	title string
}

type help struct {
	opt *CmdOpt
}

// Help 注册 help 子命令
func (opt *CmdOpt) Help(name, title, usage string) {
	h := &help{opt: opt}
	opt.New(name, title, usage, h.command)
}

func (h *help) command(fs FlagSet) DoFunc {
	return func(output io.Writer) error {
		if fs.NArg() == 0 {
			h.opt.cmd.fs.Usage()
			return nil
		}

		name := fs.Arg(0)
		for _, cmd := range h.opt.Commands() { // 调用 opt.Commands() 而不是 opt.commands，可以保证顺序一致。
			if cmd == name {
				h.opt.commands[cmd].fs.Usage()
				return nil
			}
		}

		_, err := output.Write([]byte(h.opt.notFound(name)))
		return err
	}
}

// args 表示参数列表，第一个元素为子命令名称
func (cmd *command) exec(output io.Writer, args []string) error {
	if cmd.do == nil { // 空的子命令
		return nil
	}

	if err := cmd.fs.Parse(args); err != nil {
		return err
	}
	return cmd.do(output)
}

func (opt *CmdOpt) Commands() []string {
	keys := make([]string, 0, len(opt.commands))
	for key := range opt.commands {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
