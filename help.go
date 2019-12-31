// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"fmt"
	"io"
)

type help struct {
	fs           *flag.FlagSet
	opt          *CmdOpt
	usageContent string
}

// Help 注册 help 子命令
func (opt *CmdOpt) Help(name, usage string) {
	h := &help{
		opt:          opt,
		usageContent: usage,
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

func (h *help) usage(output io.Writer) error {
	_, err := fmt.Fprint(output, h.usageContent)
	return err
}
