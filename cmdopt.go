// SPDX-License-Identifier: MIT

// Package cmdopt 用于创建子命令功能的命令行
package cmdopt

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"sort"
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
	errHandling flag.ErrorHandling
	output      io.Writer
	commands    map[string]*command

	header        string
	footer        string
	commandsTitle string
	notFound      func(string) string
	optionsTitle  string
}

// New 声明 CmdOpt 对象
//
// output 表示输出的通道；
// errHandling 表示出错时的处理方式，该值最终会被传递给子命令；
// notFound 在找不到子命令时显示的额外信息；
// header、footer、options 和 commands 作为输出帮助信息中的部分内容，
// 由用户给出。
// 帮助信息的模板如下：
//  {header}
//  {commands}:
//      cmd1    cmd1 usage
//      cmd2    cmd2 usage
//  {footer}
// 子命令的帮助信息模板如下：
//  {usage}
//  {options}:
//      -flag1    flag1 usage
//      -flag2    flag2 usage
func New(
	output io.Writer,
	errHandling flag.ErrorHandling,
	header, footer, options, commands string,
	notFound func(string) string,
) *CmdOpt {
	return &CmdOpt{
		commands:      make(map[string]*command, 10),
		errHandling:   errHandling,
		output:        output,
		header:        header,
		footer:        footer,
		optionsTitle:  options,
		commandsTitle: commands,
		notFound:      notFound,
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
func (opt *CmdOpt) New(name, usage string, do DoFunc) *flag.FlagSet {
	if _, found := opt.commands[name]; found {
		panic(fmt.Sprintf("存在相同名称的子命令：%s", name))
	}
	if usage == "" {
		panic("参数 usage 不能为空")
	}

	fs := flag.NewFlagSet(name, opt.errHandling)
	fs.SetOutput(opt.output)
	fs.Usage = func() {
		fmt.Fprint(opt.output, usage)
		if hasFlag(fs) && opt.optionsTitle != "" {
			fmt.Fprint(opt.output, "\n", opt.optionsTitle, "\n")
			origin := fs.Output()
			fs.SetOutput(opt.output)
			fs.PrintDefaults()
			fs.SetOutput(origin)
		}
	}

	var title string
	bs, err := bytes.NewBufferString(usage).ReadString('\n')
	if err == io.EOF {
		title = usage
	} else {
		title = string(bs)
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
		if _, err := opt.output.Write([]byte(opt.notFound(name))); err != nil {
			return err
		}
		return opt.usage()
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

func (opt *CmdOpt) usage() error {
	if _, err := fmt.Fprint(opt.output, opt.header); err != nil {
		return err
	}

	if opt.commandsTitle != "" && len(opt.commands) > 0 {
		if _, err := fmt.Fprint(opt.output, "\n", opt.commandsTitle, "\n"); err != nil {
			return err
		}

		for _, name := range opt.Commands() { // 保证顺序相同
			cmd := opt.commands[name]
			if _, err := fmt.Fprintf(opt.output, "%s\t%s\n", cmd.Name(), cmd.title); err != nil {
				return err
			}
		}
	}

	if opt.footer != "" {
		_, err := fmt.Fprint(opt.output, "\n", opt.footer)
		return err
	}

	return nil
}
