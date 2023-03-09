// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"io"
	"sort"
)

type command struct {
	*flag.FlagSet
	do    DoFunc
	title string
}

type help struct {
	fs  FlagSet
	opt *cmdopt
}

func (opt *cmdopt) Help(name, usage string) {
	h := &help{opt: opt}
	h.fs = opt.New(name, usage, h.do)
}

func (h *help) do(output io.Writer) error {
	if h.fs.NArg() == 0 {
		return h.opt.usage()
	}

	name := h.fs.Arg(0)
	for _, cmd := range h.opt.Commands() { // 调用 opt.Commands() 而不是 opt.commands，可以保证顺序一致。
		if cmd == name {
			h.opt.commands[cmd].Usage()
			return nil
		}
	}

	_, err := output.Write([]byte(h.opt.notFound(name)))
	return err
}

func (cmd *command) exec(output io.Writer, args []string) error {
	if err := cmd.Parse(args); err != nil {
		return err
	}
	return cmd.do(output)
}

func (opt *cmdopt) Commands() []string {
	keys := make([]string, 0, len(opt.commands))
	for key := range opt.commands {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
