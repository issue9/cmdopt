cmdopt
[![Build Status](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fissue9%2Fcmdopt%2Fbadge%3Fref%3Dmaster&style=flat)](https://actions-badge.atrox.dev/issue9/cmdopt/goto?ref=master)
[![Build Status](https://travis-ci.org/issue9/cmdopt.svg?branch=master)](https://travis-ci.org/issue9/cmdopt)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/cmdopt/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/cmdopt)
======

cmdopt 命令行选项的增强，可以轻松处理子命令。高度重用 flag 包。

```go
opt := cmdopt.New()

flag1 := opt.New("build", func(output io.Writer)error{
    output.Write([]byte("build"))
})

flag1 := opt.New("install", func(output io.Writer)error{
    output.Write([]byte("install"))
})
```

文档
----

[![Go Walker](https://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/issue9/cmdopt)
[![GoDoc](https://godoc.org/github.com/issue9/cmdopt?status.svg)](https://godoc.org/github.com/issue9/cmdopt)

版权
----

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
