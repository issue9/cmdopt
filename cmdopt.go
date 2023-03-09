// SPDX-License-Identifier: MIT

// Package cmdopt 用于创建子命令功能的命令行
package cmdopt

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// DoFunc 子命令的执行函数
type DoFunc func(io.Writer) error

// cmdopt 带子命令的命令行操作
type cmdopt struct {
	*command
	usageTemplate string
	notFound      func(string) string
	commands      map[string]*command
	maxCmdLen     int
}

// New 声明带有子命令的命令行处理对象
//
// output 表示命令行信息的输出通道；
//
// errorHandling 表示出错时的处理方式；
//
// usageTemplate 命令行的文字说明模板，包含了以下几个占位符：
//   - {{flags}} 参数说明，输出时被参数替换，如果没有可以为空；
//   - {{commands}} 子命令说明，输出时被子命令列表替换，如果没有可以为空；
//
// notFound 表示找不到子命令时需要返回的文字说明，若为空，则采用 usageTemplate 处理后的内容；
func New(output io.Writer, errorHandling flag.ErrorHandling, usageTemplate string, do DoFunc, notFound func(string) string) CmdOpt {
	fs := flag.NewFlagSet("", errorHandling)
	fs.SetOutput(output)

	return &cmdopt{
		command:       &command{do: do, FlagSet: fs},
		usageTemplate: usageTemplate,
		notFound:      notFound,
		commands:      make(map[string]*command, 10),
	}
}

func (opt *cmdopt) New(name, usage string, do DoFunc) FlagSet {
	return opt.Add(flag.NewFlagSet(name, opt.ErrorHandling()), do, usage)
}

func (opt *cmdopt) Add(fs *flag.FlagSet, do DoFunc, usage string) FlagSet {
	name := fs.Name()
	if _, found := opt.commands[name]; found {
		panic(fmt.Sprintf("存在相同名称的子命令：%s", name))
	}
	if usage == "" {
		panic("参数 usage 不能为空")
	}

	var title string
	bs, err := bytes.NewBufferString(usage).ReadString('\n')
	if errors.Is(err, io.EOF) {
		title = usage
	} else {
		title = bs[:len(bs)-1] // 去掉换行符
	}
	if strings.Contains(title, "{{flags}}") {
		panic("usage 第一行中不能包含 {{flags}}")
	}

	fs.Init(name, opt.ErrorHandling())
	fs.SetOutput(opt.Output())
	fs.Usage = func() {
		fmt.Fprintln(opt.Output(), strings.ReplaceAll(usage, "{{flags}}", getFlags(fs)))
	}

	opt.commands[name] = &command{
		FlagSet: fs,
		do:      do,
		title:   title,
	}

	if l := len(name); l > opt.maxCmdLen {
		opt.maxCmdLen = l
	}

	return fs
}

func getFlags(fs *flag.FlagSet) string {
	var bs bytes.Buffer
	old := fs.Output()
	fs.SetOutput(&bs)
	fs.PrintDefaults()
	fs.SetOutput(old)
	return bs.String()
}

func (opt *cmdopt) Exec(args []string) error {
	// NOTE: 让用户提供参数，而不是直接产从 os.Args 中取，
	// 可以方便用户作一些调试操作。

	if len(args) == 0 {
		return opt.command.exec(opt.Output(), nil)
	}

	name := args[0]
	if name[0] == '-' { // 非子命令模式
		return opt.command.exec(opt.Output(), args)
	}

	if cmd, found := opt.commands[name]; found {
		return cmd.exec(opt.Output(), args[1:])
	}

	if opt.notFound != nil {
		_, err := io.WriteString(opt.Output(), opt.notFound(name))
		return err
	}
	return opt.usage()
}

func (opt *cmdopt) usage() error {
	flags := getFlags(opt.FlagSet)
	var commands bytes.Buffer
	for _, name := range opt.Commands() { // 保证顺序相同
		cmd := opt.commands[name]
		cmdName := name + strings.Repeat(" ", opt.maxCmdLen+3-len(name)) // 为子命令名称留下的最小长度
		if _, err := fmt.Fprintf(&commands, "  %s%s\n", cmdName, cmd.title); err != nil {
			return err
		}
	}

	usage := strings.ReplaceAll(opt.usageTemplate, "{{flags}}", flags)
	usage = strings.ReplaceAll(usage, "{{commands}}", commands.String())
	_, err := fmt.Fprintln(opt.Output(), usage)
	return err
}
