// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"io"
)

type help struct {
	fs    *flag.FlagSet
	opt   *CmdOpt
	usage DoFunc
}

// Help 注册 help 子命令
func (opt *CmdOpt) Help(name string, usage DoFunc) {
	h := &help{
		opt:   opt,
		usage: usage,
	}
	h.fs = opt.New(name, h.do, h.usage)
}

func (h *help) do(output io.Writer) error {
	if h.fs.NArg() == 0 {
		return h.usage(output)
	}

	name := h.fs.Arg(1)
	for k, v := range h.opt.commands {
		if k == name {
			v.Usage()
			return nil
		}
	}

	if _, err := output.Write([]byte(h.opt.notFound(name))); err != nil {
		return err
	}
	return h.opt.usage(output)
}
