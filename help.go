// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Help 注册 help 子命令
//
// 注册之后可通过 `help <name>` 来获取该子命令的使用说明。
func Help(opt *CmdOpt, name, title, usage string) {
	f := func(fs *flag.FlagSet) DoFunc {
		return func(output io.Writer) error {
			if fs.NArg() == 0 {
				_, err := io.WriteString(output, opt.usage())
				return err
			}

			name := fs.Arg(0)
			if _, usage, found := opt.Command(name); found {
				_, err := io.WriteString(output, usage)
				return err
			}

			_, err := io.WriteString(output, opt.notFound(name))
			return err
		}
	}

	opt.New(name, title, usage, f)
}
