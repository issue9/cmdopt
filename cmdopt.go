// SPDX-License-Identifier: MIT

// Package cmdopt 用于创建子命令功能的命令行
package cmdopt

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
)

// DoFunc 子命令的执行函数
type DoFunc func(io.Writer) error

type command struct {
	*flag.FlagSet
	do    DoFunc
	title string
}

// CmdOpt 带子命令的命令行操作
type CmdOpt struct {
	// 表示输出的通道
	//
	// 此值必须指定；
	Output io.Writer

	// 表示出错时的处理方式
	//
	// 该值最终会被传递给子命令； 为空则采用 flag.ContinueOnError
	ErrorHandling flag.ErrorHandling

	// Header、Footer、OptionsTitle 和 CommandsTitle 作为输出帮助信息中的部分内容
	//
	// 帮助信息的模板如下：
	//  {Header}
	//  {CommandsTitle}:
	//      cmd1    cmd1 usage
	//      cmd2    cmd2 usage
	//  {Footer}
	//
	// 子命令的帮助信息模板如下：
	//  usage
	//  {OptionsTitle}:
	//      -flag1    flag1 usage
	//      -flag2    flag2 usage
	Header        string
	Footer        string
	CommandsTitle string
	OptionsTitle  string

	// 在找不到子命令时显示的额外信息
	//
	// 其中参数为子命令的名称。
	NotFound func(string) string

	commands map[string]*command
}

// New 注册一条新的子命令
//
// name 为子命令的名称，必须唯一；
// do 为该条子命令执行的函数体；
// usage 为该条子命令的帮助内容输出，当 usage 为多行是，其第一行作为此命令的摘要信息。
//
// 返回 FlagSet，不需要手动调用 FlagSet.Parse，该方法会在执行时自动执行。
// FlagSet.Args 返回的是包含了子命令在内容的所有内容。
func (opt *CmdOpt) New(name, usage string, do DoFunc) *flag.FlagSet {
	if opt.commands == nil {
		opt.commands = make(map[string]*command, 10)
	}

	if _, found := opt.commands[name]; found {
		panic(fmt.Sprintf("存在相同名称的子命令：%s", name))
	}
	if usage == "" {
		panic("参数 usage 不能为空")
	}

	fs := flag.NewFlagSet(name, opt.ErrorHandling)
	fs.SetOutput(opt.Output)
	fs.Usage = func() {
		fmt.Fprint(opt.Output, usage)
		if hasFlag(fs) && opt.OptionsTitle != "" {
			fmt.Fprint(opt.Output, "\n", opt.OptionsTitle, "\n")
			origin := fs.Output()
			fs.SetOutput(opt.Output)
			fs.PrintDefaults()
			fs.SetOutput(origin)
		}
	}

	var title string
	bs, err := bytes.NewBufferString(usage).ReadString('\n')
	if errors.Is(err, io.EOF) {
		title = usage
	} else {
		title = bs
		title = title[:len(title)-1]
	}
	opt.commands[name] = &command{
		FlagSet: fs,
		do:      do,
		title:   title,
	}

	return fs
}

func hasFlag(fs *flag.FlagSet) bool {
	var has bool
	fs.VisitAll(func(*flag.Flag) {
		has = true
	})
	return has
}

// Exec 执行命令行程序
//
// args 第一个元素应该是子命令名称。
func (opt *CmdOpt) Exec(args []string) error {
	if len(args) == 0 {
		return opt.usage()
	}

	name := args[0]
	args = args[1:]

	cmd, found := opt.commands[name]
	if !found {
		if opt.NotFound != nil {
			if _, err := opt.Output.Write([]byte(opt.NotFound(name))); err != nil {
				return err
			}
		}
		return opt.usage()
	}

	if err := cmd.Parse(args); err != nil {
		return err
	}

	return cmd.do(opt.Output)
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

func (opt *CmdOpt) usage() error {
	if _, err := fmt.Fprint(opt.Output, opt.Header); err != nil {
		return err
	}

	if opt.CommandsTitle != "" && len(opt.commands) > 0 {
		if _, err := fmt.Fprint(opt.Output, "\n", opt.CommandsTitle, "\n"); err != nil {
			return err
		}

		cmds := opt.Commands()
		var max int
		for _, cmd := range cmds {
			if len(cmd) > max {
				max = len(cmd)
			}
		}
		max += 3

		for _, name := range cmds { // 保证顺序相同
			cmd := opt.commands[name]
			cmdName := name + strings.Repeat(" ", max-len(name))
			if _, err := fmt.Fprintf(opt.Output, "    %s%s\n", cmdName, cmd.title); err != nil {
				return err
			}
		}
	}

	if opt.Footer != "" {
		_, err := fmt.Fprint(opt.Output, "\n", opt.Footer)
		return err
	}

	return nil
}
