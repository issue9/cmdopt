cmdopt
[![Go](https://github.com/issue9/cmdopt/workflows/Go/badge.svg)](https://github.com/issue9/cmdopt/actions?query=workflow%3AGo)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/cmdopt/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/cmdopt)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/cmdopt)](https://pkg.go.dev/github.com/issue9/cmdopt)
![Go version](https://img.shields.io/github/go-mod/go-version/issue9/cmdopt)
======

cmdopt 命令行选项的增强，可以轻松处理子命令。高度重用 flag 包。

```go
opt := &cmdopt.New(...)

// 子命令 build，为一个 flag.FlagSet 实例
build := opt.New("build", "title", "usage", func(f *flag.FlagSet) DoFunc {
    v := f.Bool("v", "false", ...)

    return func(output io.Writer)error{
        if v {
            ...
        } else {
            output.Write([]byte("build"))
        }
    }
})

// 子命令 install
install := opt.New("install", "title", "usage", func(*flag.FlagSet) DoFunc {
    return func(output io.Writer)error{
        output.Write([]byte("install"))
    }
})
```

版权
----

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
