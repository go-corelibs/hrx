[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/go-corelibs/hrx)
[![codecov](https://codecov.io/gh/go-corelibs/hrx/graph/badge.svg?token=aaJBn82EzN)](https://codecov.io/gh/go-corelibs/hrx)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-corelibs/hrx)](https://goreportcard.com/report/github.com/go-corelibs/hrx)

# hrx - human readable archive utilities

A collection of utilities for working with `.hrx` files.

# Installation

``` shell
> go get github.com/go-corelibs/hrx@latest
```

# Examples

## Archive

``` go
func main() {
    a := hrx.New("", "file-level comment")
    _ = a.Set("file.txt", "this is a simple test file\nwith multiple lines")
    if err := a.WriteFile("/path/to/new.hrx"); err != nil {
        panic(err)
    }
    // there is now a /path/to/new.hrx with one file.txt archived
}
```

# Go-CoreLibs

[Go-CoreLibs] is a repository of shared code between the [Go-Curses] and
[Go-Enjin] projects.

# License

```
Copyright 2024 The Go-CoreLibs Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use file except in compliance with the License.
You may obtain a copy of the license at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

[Go-CoreLibs]: https://github.com/go-corelibs
[Go-Curses]: https://github.com/go-curses
[Go-Enjin]: https://github.com/go-enjin
