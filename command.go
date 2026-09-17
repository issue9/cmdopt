// SPDX-FileCopyrightText: 2019-2026 caixw
//
// SPDX-License-Identifier: MIT

package cmdopt

import (
	"flag"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// 对子命令的描述
//
// 与 [CmdOpt.New] 的参数一一对应
type Command struct {
	// 子命令名称
	Name string

	// 简短描述
	Title string

	// 详细说明
	Usage string

	// 注册的执行方法
	Command CommandFunc
}

// CommandFunc 子命令的初始化方法
//
// FlagSet 可用于绑定各个命令行参数；
// 返回值 [DoFunc] 表示实际执行的函数；
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

// New 注册一条新的子命令
//
// name 为子命令的名称，必须唯一；
// title 子命令的简要说明，不能包含换行符；
// usage 为该条子命令的帮助内容。可以包含 {{flags}} 占位符，表示参数信息。
// cmd 为该条子命令执行的函数体，具体可参考 [CommandFunc]；
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

	usage = strings.ReplaceAll(usage, "{{flags}}", getFlagSetUsage(fs))
	if usage[len(usage)-1] != '\n' {
		usage += "\n"
	}

	fs.Usage = func() { io.WriteString(opt.Output(), usage) }

	opt.NewPlain(name, title, usage, do2exec(do, fs))
}

// NewCommand 注册新的子命令
//
// 功能与 [CmdOpt.New] 完全相同，但是可以添加多个。
func (opt *CmdOpt) NewCommand(cmd ...*Command) {
	for _, c := range cmd {
		opt.New(c.Name, c.Title, c.Usage, c.Command)
	}
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

// Commands 返回所有的子命令
func (opt *CmdOpt) Commands() []string {
	keys := slices.Collect(maps.Keys(opt.commands))
	slices.Sort(keys)
	return keys
}

// Command 返回指定的命令的说明
func (opt *CmdOpt) Command(name string) (title, usage string, found bool) {
	if cmd, found := opt.commands[name]; found {
		return cmd.title, cmd.usage, true
	}
	return "", "", false
}
