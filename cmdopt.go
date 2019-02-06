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
	ErrHandling flag.ErrorHandling
	Output      io.Writer // 输出通道
	Usage       func()
}

// New 声明一条新的子命令
func (opt *CmdOpt) New(name string, do DoFunc) *flag.FlagSet {
	if opt.commands == nil {
		opt.commands = make(map[string]*command, 10)
	}

	if _, found := opt.commands[name]; found {
		panic("存在相同名称的数据")
	}

	fs := flag.NewFlagSet(name, opt.ErrHandling)

	opt.commands[name] = &command{
		FlagSet: fs,
		do:      do,
	}

	return fs
}

// Exec 执行命令行程序
func (opt *CmdOpt) Exec(args []string) error {
	if len(args) == 0 {
		opt.Usage()
		return nil
	}

	cmd, found := opt.commands[args[1]]
	if !found {
		return ErrNotFound
	}

	if err := cmd.Parse(args[2:]); err != nil {
		return err
	}

	return cmd.do(opt.Output)
}

// Help 注释 cmd help xxx 的命令，其中子命令 help 通过 name 指定
func (opt *CmdOpt) Help(name string) error {
	for k, v := range opt.commands {
		if k == name {
			v.Usage()
			return nil
		}
	}
	return nil
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
