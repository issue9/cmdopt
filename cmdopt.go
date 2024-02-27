// SPDX-FileCopyrightText: 2019-2024 caixw
//
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

// CmdOpt 带子命令的命令行操作
type CmdOpt struct {
	cmd *command
	// 生成整个整个命令行的使用说明
	usage func() string

	output      io.Writer
	errHandling flag.ErrorHandling
	notFound    func(string) string
	commands    map[string]*command
	maxCmdLen   int // 记录子命令的最大字符宽度，使输出的命令行可以更加美观。

	execed bool
}

// CommandFunc 子命令的初始化方法
//
// FlagSet 可用于绑定各个命令行参数；
// 返回值 [DoFunc] 表示实际执行的函数；
//
// 一般与 [DoFunc] 组合使用：
//
//	func(fs *flag.FlagSet) DoFunc {
//	    f1 := fs.Bool("f1", true, "usage")
//	    return func(w io.Writer) error {
//	        if *f1 { TODO }
//	    }
//	}
//
// 在 CommandFunc 中初始化 flag 参数，并在其返回函数中作实际处理，这样可以防止大量的全局变量的声明。
//
// 如非必要情况，CommandFunc 的 FlagSet 只用于绑定参数，不应该修改其相关配置。
type CommandFunc = func(*flag.FlagSet) DoFunc

// DoFunc 命令行的实际执行方法
//
// io.Writer 用于内容的输出，如果有错误信息应该通过返回值返回。
type DoFunc = func(io.Writer) error

// New 声明带有子命令的命令行处理对象
//
// output 表示命令行信息的输出通道；
// errorHandling 表示出错时的处理方式；
// cmd 非子命令的参数设定，可以为空；
// usageTemplate 命令行的文字说明模板；
// notFound 表示找不到子命令时需要返回的文字说明，若为空，则采用 usageTemplate 处理后的内容；
//
// usageTemplate 可以包含了以下几个占位符：
//   - {{flags}} 参数说明，输出时被参数替换，如果没有可以为空；
//   - {{commands}} 子命令说明，输出时被子命令列表替换，如果没有可以为空；
func New(output io.Writer, errorHandling flag.ErrorHandling, usageTemplate string, cmd CommandFunc, notFound func(string) string) *CmdOpt {
	fs := flag.NewFlagSet("", errorHandling)
	fs.SetOutput(output)

	do := func(w io.Writer) error { return nil }
	if cmd != nil {
		do = cmd(fs)
	}

	opt := &CmdOpt{
		cmd: &command{exec: do2exec(do, fs)},

		output:      output,
		errHandling: errorHandling,
		notFound:    notFound,
		commands:    make(map[string]*command, 10),
	}

	opt.usage = func() string {
		opt.buildUsage(usageTemplate, fs)
		return opt.cmd.usage
	}

	fs.Usage = func() {
		io.WriteString(opt.Output(), opt.usage())
	}

	return opt
}

// New 注册一条新的子命令
//
// name 为子命令的名称，必须唯一；
// cmd 为该条子命令执行的函数体，具体可参考 [CommandFunc]；
// usage 为该条子命令的帮助内容。可以包含 {{flags}} 占位符，表示参数信息。
func (opt *CmdOpt) New(name, title, usage string, cmd CommandFunc) {
	if name == "" {
		panic("参数 name 不能为空")
	}
	if usage == "" {
		panic("参数 usage 不能为空")
	}
	if cmd == nil {
		panic("参数 cmd 不能为空")
	}
	if _, found := opt.commands[name]; found {
		panic(fmt.Sprintf("存在相同名称的子命令：%s", name))
	}

	fs := flag.NewFlagSet(name, opt.errHandling)
	fs.SetOutput(opt.output)
	do := cmd(fs) // 确定 flag，需要在生成 usage 之前调用

	usage = strings.ReplaceAll(usage, "{{flags}}", getFlags(fs))
	if usage[len(usage)-1] != '\n' {
		usage += "\n"
	}

	fs.Usage = func() { io.WriteString(opt.Output(), usage) }

	opt.NewPlain(name, title, usage, do2exec(do, fs))
}

func do2exec(do DoFunc, fs *flag.FlagSet) func(io.Writer, []string) error {
	return func(w io.Writer, args []string) error {
		if err := fs.Parse(args); err != nil {
			return err
		}
		return do(w)
	}
}

// NewPlain 添加自行处理参数的子命令
//
// 用户需要在 exec 中自行处理命令行参数，exec 原型如下：
//
//	func(output io.Writer, args []string) error
//
// output 即为 [CmdOpt.Output]，args 为子命令的参数，不包含子命令本身。
//
// name, title 和 usage 参数可参考 [CmdOpt.New]，唯一不同点是 usage 不会处理 {{flags}} 占位符。
func (opt *CmdOpt) NewPlain(name, title, usage string, exec func(io.Writer, []string) error) {
	if opt.execed {
		panic("程序已经运行，不可再添加子命令！")
	}

	opt.commands[name] = &command{
		exec:  exec,
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
		return opt.cmd.exec(opt.Output(), nil)
	}

	name := args[0]
	if name[0] == '-' { // 第一个即为参数，表示为非子命令模式
		if err := opt.cmd.exec(opt.Output(), args); err != nil && !errors.Is(err, flag.ErrHelp) {
			return err
		}
		return nil
	}

	if cmd, found := opt.commands[name]; found {
		return cmd.exec(opt.Output(), args[1:])
	}

	if opt.notFound != nil {
		_, err := io.WriteString(opt.Output(), opt.notFound(name))
		return err
	}

	_, err := io.WriteString(opt.Output(), opt.Usage())
	return err
}

// Usage 整个项目的使用说明内容
//
// 基于 [New] 的 usage 参数，里面的占位符会被真实的内容所覆盖。
// 每次调用时都根据当前的命令行情况重新生成内容。
func (opt *CmdOpt) Usage() string { return opt.usage() }

func (opt *CmdOpt) buildUsage(tpl string, fs *flag.FlagSet) {
	flags := getFlags(fs)
	var commands bytes.Buffer
	for _, name := range opt.Commands() { // 保证顺序相同
		title, _, _ := opt.Command(name)
		cmdName := name + strings.Repeat(" ", opt.maxCmdLen+3-len(name)) // 为子命令名称留下的最小长度
		fmt.Fprintf(&commands, "  %s%s\n", cmdName, title)
	}

	usage := strings.ReplaceAll(tpl, "{{flags}}", flags)
	usage = strings.ReplaceAll(usage, "{{commands}}", commands.String())

	if len(usage) > 0 && usage[len(usage)-1] != '\n' {
		usage += "\n"
	}

	opt.cmd.usage = usage
}

// SetOutput 设置输出通道
func (opt *CmdOpt) SetOutput(w io.Writer) { opt.output = w }

func (opt *CmdOpt) Output() io.Writer { return opt.output }
