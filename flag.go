// SPDX-License-Identifier: MIT

package cmdopt

import (
	"encoding"
	"flag"
	"io"
	"time"
)

// FlagSet 这是 [flag.FlagSet] 的子集
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
