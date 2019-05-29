// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdopt

import (
	"flag"
	"fmt"
	"io"
)

type help struct {
	fs  *flag.FlagSet
	opt *CmdOpt
}

// Help 注册 help 子命令
func (opt *CmdOpt) Help(name string) {
	h := &help{opt: opt}
	h.fs = opt.New(name, h.do, h.usage)
}

func (h *help) do(output io.Writer) error {
	if len(h.fs.Args()) == 1 {
		_, err := fmt.Fprintln(output, "未指定查询的命令名称")
		if err != nil {
			return err
		}

		return h.usage(output)
	}
	name := h.fs.Arg(1)
	for k, v := range h.opt.commands {
		if k == name {
			v.Usage()
			return nil
		}
	}

	if _, err := h.opt.output.Write(notFound(name)); err != nil {
		return err
	}
	return h.opt.usage(h.opt.output)
}

func (h *help) usage(output io.Writer) error {
	_, err := fmt.Fprintln(output, `查看各个子命令的帮助内容`)
	return err
}
