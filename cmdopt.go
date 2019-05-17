// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cmdopt 命令行选项功能
package cmdopt

import (
	"errors"
	"flag"
	"io"
	"sort"
)

// ErrNotFound 子命令不存在时，返回该错误信息
var ErrNotFound = errors.New("不存在的子命令")

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
	output      io.Writer // 输出通道
	usage       func(io.Writer)
}

// New 新的 CmdOpt 对象
func New(output io.Writer, errHandling flag.ErrorHandling, usage func(io.Writer)) *CmdOpt {
	return &CmdOpt{
		commands:    make(map[string]*command, 10),
		errHandling: errHandling,
		output:      output,
		usage:       usage,
	}
}

// New 注册一条新的子命令
//
// name 为子命令的名称，必须唯一；
// do 为该条子命令执行的函数体。
//
// 返回 FlagSet，不需要手动调用 FlagSet.Parse，
// 该方法会在执行时自动执行，传递给 FlagSet.Parse() 的参数中为 os.Args[2:]
func (opt *CmdOpt) New(name string, do DoFunc) *flag.FlagSet {
	if _, found := opt.commands[name]; found {
		panic("存在相同名称的数据")
	}

	fs := flag.NewFlagSet(name, opt.errHandling)
	fs.SetOutput(opt.output)

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
		opt.usage(opt.output)
		return nil
	}

	cmd, found := opt.commands[args[0]]
	if !found {
		return ErrNotFound
	}

	if err := cmd.Parse(args[1:]); err != nil {
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
