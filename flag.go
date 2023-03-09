// SPDX-License-Identifier: MIT

package cmdopt

import (
	"encoding"
	"flag"
	"io"
	"time"
)

type CmdOpt interface {
	FlagSet

	// New 注册一条新的子命令
	//
	// name 为子命令的名称，必须唯一；
	// do 为该条子命令执行的函数体；
	// usage 为该条子命令的帮助内容输出。
	// 当 usage 为多行是，其第一行作为此命令的摘要信息。可以包含 {{flags}} 占位符，表示输出参数信息，但是不能在第一行中出现。
	New(name, usage string, do DoFunc) FlagSet

	// Add 添加一条新的子命令
	//
	// 参数说明可参考 [CmdOpt.New]。
	// 子命令的名称根据 fs.Name 获取。
	// NOTE: 这会托管 fs 的 Output、ErrorHandling 以及 Usage 对象。
	Add(fs *flag.FlagSet, do DoFunc, usage string) FlagSet

	// Commands 所有的子命令列表
	Commands() []string

	// Exec 执行命令行程序
	//
	// args 忽略程序名之后的参数列表，比如 os.Args[1:]。
	Exec(args []string) error

	// Help 注册 help 子命令
	Help(name, usage string)
}

// FlagSet 子命令操作返回的接口
type FlagSet interface {
	Arg(i int) string
	Args() []string
	Bool(name string, value bool, usage string) *bool
	BoolVar(p *bool, name string, value bool, usage string)
	Duration(name string, value time.Duration, usage string) *time.Duration
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
	Float64(name string, value float64, usage string) *float64
	Float64Var(p *float64, name string, value float64, usage string)
	Func(name, usage string, fn func(string) error)
	Int(name string, value int, usage string) *int
	Int64(name string, value int64, usage string) *int64
	Int64Var(p *int64, name string, value int64, usage string)
	IntVar(p *int, name string, value int, usage string)
	Lookup(name string) *flag.Flag
	NArg() int
	NFlag() int
	Name() string
	Output() io.Writer
	Set(name, value string) error
	String(name string, value string, usage string) *string
	StringVar(p *string, name string, value string, usage string)
	TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string)
	Uint(name string, value uint, usage string) *uint
	Uint64(name string, value uint64, usage string) *uint64
	Uint64Var(p *uint64, name string, value uint64, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Var(value flag.Value, name string, usage string)
	Visit(fn func(*flag.Flag))
	VisitAll(fn func(*flag.Flag))
}
