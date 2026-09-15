// SPDX-FileCopyrightText: 2019-2026 caixw
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

type command struct {
	exec  func(io.Writer, []string) error
	title string
	usage string
}

// CmdOpt 支持子命令的命令行操作
type CmdOpt struct {
	rootCmd *command

	usage func() string // 生成整个命令行的使用说明

	output      io.Writer
	errHandling flag.ErrorHandling
	notFound    func(string) string
	commands    map[string]*command
	maxCmdLen   int // 记录子命令的最大字符宽度，使输出的命令行可以更加美观。

	execed bool
}

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
	rootFS := flag.NewFlagSet("", errorHandling)
	rootFS.SetOutput(output)

	do := func(w io.Writer) error { return nil }
	if cmd != nil {
		do = cmd(rootFS)
	}

	opt := &CmdOpt{
		rootCmd: &command{exec: do2exec(do, rootFS)},

		output:      output,
		errHandling: errorHandling,
		notFound:    notFound,
		commands:    make(map[string]*command, 10),
	}

	opt.usage = func() string {
		opt.buildUsage(usageTemplate, rootFS)
		return opt.rootCmd.usage
	}

	rootFS.Usage = func() {
		io.WriteString(opt.Output(), opt.usage())
	}

	return opt
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
		return opt.rootCmd.exec(opt.Output(), nil)
	}

	name := args[0]
	if name[0] != '-' { // 非 - 开头，先尝试是否为子命令，如果找不到再执行主命令。
		if cmd, found := opt.commands[name]; found {
			return cmd.exec(opt.Output(), args[1:])
		}
	}

	err := opt.rootCmd.exec(opt.Output(), args)
	if errors.Is(err, flag.ErrHelp) {
		_, err = io.WriteString(opt.Output(), opt.Usage())
	}

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

	opt.rootCmd.usage = usage
}

// SetOutput 设置输出通道
func (opt *CmdOpt) SetOutput(w io.Writer) { opt.output = w }

func (opt *CmdOpt) Output() io.Writer { return opt.output }

func (opt *CmdOpt) ErrorHandling() flag.ErrorHandling { return opt.errHandling }
