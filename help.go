// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdopt

import (
	"flag"
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
	for k, v := range h.opt.commands {
		if k == h.fs.Arg(0) {
			v.Usage()
			return nil
		}
	}

	h.opt.output.Write(notFound)
	h.opt.usage(h.opt.output)
	return nil
}

func (h *help) usage(output io.Writer) error {
	_, err := output.Write([]byte("查看各个子命令的帮助内容"))
	return err
}
