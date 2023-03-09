// SPDX-License-Identifier: MIT

package cmdopt

import "io"

type help struct {
	fs  FlagSet
	opt *CmdOpt
}

// Help 注册 help 子命令
func (opt *CmdOpt) Help(name, usage string) {
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

	_, err := output.Write([]byte(h.opt.NotFound(name)))
	return err
}
