// SPDX-License-Identifier: MIT

// Package cmdopt 用于创子命令功能的命令行。
package cmdopt

import (
	"flag"
	"io"
	"sort"
)

// DoFunc 子命令的执行函数
type DoFunc func(io.Writer) error

type command struct {
	*flag.FlagSet
	do DoFunc
}

// CmdOpt 带子命令的命令行操作
type CmdOpt struct {
	commands    map[string]*command
	errHandling flag.ErrorHandling
	output      io.Writer
	usage       DoFunc
	notFound    func(string) string
}

// New 新的 CmdOpt 对象
//
// output 表示命令输出内容所指向的通道；
// errHandling 传递给每一个 FlagSet 的内容；
// usage 表示命令行的默认显示内容，在未指定子命令或是未找到子命令时会显示此内容；
// notFound 未找到子命令时，会额外显示此内容。
func New(output io.Writer, errHandling flag.ErrorHandling, usage DoFunc, notFound func(string) string) *CmdOpt {
	return &CmdOpt{
		commands:    make(map[string]*command, 10),
		errHandling: errHandling,
		output:      output,
		usage:       usage,
		notFound:    notFound,
	}
}

// New 注册一条新的子命令
//
// name 为子命令的名称，必须唯一；
// do 为该条子命令执行的函数体；
// usage 为该条子命令的帮助内容输出。
// 如果 usage 为 nil，则采用自带的内容，也可以通过返回值再次指定。
//
// 返回 FlagSet，不需要手动调用 FlagSet.Parse，该方法会在执行时自动执行。
// FlagSet.Args 返回的是包含了子命令在内容的所有内容。
func (opt *CmdOpt) New(name string, do, usage DoFunc) *flag.FlagSet {
	if _, found := opt.commands[name]; found {
		panic("存在相同名称的数据")
	}

	fs := flag.NewFlagSet(name, opt.errHandling)
	fs.SetOutput(opt.output)

	if usage != nil {
		fs.Usage = func() {
			usage(fs.Output())
		}
	}

	opt.commands[name] = &command{
		FlagSet: fs,
		do:      do,
	}

	return fs
}

// Exec 执行命令行程序
//
// args 第一个元素应该是子命令名称。
func (opt *CmdOpt) Exec(args []string) error {
	if len(args) == 0 {
		return opt.usage(opt.output)
	}

	cmd, found := opt.commands[args[0]]
	if !found {
		notFound := []byte(opt.notFound(args[0]))
		if _, err := opt.output.Write(notFound); err != nil {
			return err
		}
		return opt.usage(opt.output)
	}

	if err := cmd.Parse(args); err != nil {
		return err
	}

	return cmd.do(opt.output)
}

// Commands 所有的子命令列表
func (opt *CmdOpt) Commands() []string {
	keys := make([]string, 0, len(opt.commands))

	for key := range opt.commands {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
