// SPDX-License-Identifier: MIT

// Package cmdopt 用于创建子命令功能的命令行
package cmdopt

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"
)

// CmdOpt 带子命令的命令行操作
type CmdOpt struct {
	cmd           *command
	usageTemplate string
	notFound      func(string) string
	commands      map[string]*command
	maxCmdLen     int

	execed bool
}

// CommandFunc 子命令的初始化方法
//
// FlagSet 可用于绑定各个命令行参数；
// 返回值 [DoFunc] 表示实际执行的函数；
type CommandFunc = func(FlagSet) DoFunc

type DoFunc = func(io.Writer) error

// New 声明带有子命令的命令行处理对象
//
// output 表示命令行信息的输出通道；
//
// errorHandling 表示出错时的处理方式；
//
// cmd 非子命令的参数设定，可以为空；
//
// usageTemplate 命令行的文字说明模板，包含了以下几个占位符：
//   - {{flags}} 参数说明，输出时被参数替换，如果没有可以为空；
//   - {{commands}} 子命令说明，输出时被子命令列表替换，如果没有可以为空；
//
// notFound 表示找不到子命令时需要返回的文字说明，若为空，则采用 usageTemplate 处理后的内容；
func New(output io.Writer, errorHandling flag.ErrorHandling, usageTemplate string, cmd CommandFunc, notFound func(string) string) *CmdOpt {
	fs := flag.NewFlagSet("", errorHandling)
	fs.SetOutput(output)

	var do DoFunc
	if cmd != nil {
		do = cmd(fs)
	}

	opt := &CmdOpt{
		cmd:           &command{do: do, fs: fs},
		usageTemplate: usageTemplate,
		notFound:      notFound,
		commands:      make(map[string]*command, 10),
	}

	fs.Usage = func() {
		fmt.Fprintln(opt.cmd.fs.Output(), opt.usage())
	}

	return opt
}

// New 注册一条新的子命令
//
// name 为子命令的名称，必须唯一；
// cmd 为该条子命令执行的函数体；
// usage 为该条子命令的帮助内容。可以包含 {{flags}} 占位符，表示参数信息。
func (opt *CmdOpt) New(name, title, usage string, cmd CommandFunc) {
	fs := flag.NewFlagSet(name, opt.cmd.fs.ErrorHandling())

	if _, found := opt.commands[name]; found {
		panic(fmt.Sprintf("存在相同名称的子命令：%s", name))
	}
	if usage == "" {
		panic("参数 usage 不能为空")
	}
	usage = strings.ReplaceAll(usage, "{{flags}}", getFlags(fs))

	fs.Init(name, opt.cmd.fs.ErrorHandling())
	fs.SetOutput(opt.cmd.fs.Output())
	fs.Usage = func() {
		fmt.Fprintln(opt.cmd.fs.Output(), usage)
	}

	opt.commands[name] = &command{
		fs:    fs,
		do:    cmd(fs),
		title: title,
		usage: usage,
	}

	if l := len(name); l > opt.maxCmdLen {
		opt.maxCmdLen = l
	}
}

func getFlags(fs *flag.FlagSet) string {
	var bs bytes.Buffer
	old := fs.Output()
	fs.SetOutput(&bs)
	fs.PrintDefaults()
	fs.SetOutput(old)
	return bs.String()
}

// Exec 执行命令行程序
//
// args 参数列表，不包含应用名称，比如 os.Args[1:]。
func (opt *CmdOpt) Exec(args []string) error {
	// NOTE: 让用户提供参数，而不是直接产从 os.Args 中取，可以方便用户作一些调试操作。

	if opt.execed {
		panic("不可多次调用 Exec 方法")
	}
	opt.execed = true

	if len(args) == 0 {
		return opt.cmd.exec(nil)
	}

	name := args[0]
	if name[0] == '-' { // 非子命令模式
		return opt.cmd.exec(args)
	}

	if cmd, found := opt.commands[name]; found {
		return cmd.exec(args[1:])
	}

	if opt.notFound != nil {
		_, err := io.WriteString(opt.cmd.fs.Output(), opt.notFound(name))
		return err
	}

	opt.cmd.fs.Usage()
	return nil
}

func (opt *CmdOpt) usage() string {
	flags := getFlags(opt.cmd.fs)
	var commands bytes.Buffer
	for _, name := range opt.Commands() { // 保证顺序相同
		cmd := opt.commands[name]
		cmdName := name + strings.Repeat(" ", opt.maxCmdLen+3-len(name)) // 为子命令名称留下的最小长度
		fmt.Fprintf(&commands, "  %s%s\n", cmdName, cmd.title)
	}

	usage := strings.ReplaceAll(opt.usageTemplate, "{{flags}}", flags)
	usage = strings.ReplaceAll(usage, "{{commands}}", commands.String())

	return usage
}
