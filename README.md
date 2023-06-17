<!-- markdownlint-disable MD041 -->
[![go1.18+](https://img.shields.io/badge/Go-1.18+-blue?logo=go)](https://github.com/KEINOS/go-md-spec-check/blob/main/.github/workflows/unit-tests.yml#L81 "Supported versions")
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-md-spec-check.svg)](https://pkg.go.dev/github.com/KEINOS/go-md-spec-check/ "View document online")

# Markdown Specification Checker for Go

[go-md-spec-check](https://github.com/KEINOS/go-md-spec-check) is a simple Go package to **check if a function can convert Markdown to HTML according to the [CommonMark](https://commonmark.org/) specification**.

## Usage

```go
// Download module (go 1.18+)
go get github.com/KEINOS/go-md-spec-check
```

```go
// Import package
import "github.com/KEINOS/go-md-spec-check/mdspec"
```

In the below example, `mdspec.SpecCheck()` runs `myMarkdownParser()` against about 500-600 test cases over the CommonMark v0.30 specification. And returns the first error encountered that does not comply with the CommonMark specification.

```go
import (
  "fmt"
  "log"

  "github.com/KEINOS/go-md-spec-check/mdspec"
)

func Example() {
    // Sample Markdown-to-HTML conversion function that does not do its job.
    myMarkdownParser := func(markdown string) (string, error) {
        return "<p>Hello, World!</p>", nil
    }

    // Check if the `myMarkdownParser()` complies with the CommonMark specification
    // version 0.30.
    // Choices: "v0.13", "v0.14" ... "v0.30"
    err := mdspec.SpecCheck("v0.30", myMarkdownParser)

    if err != nil {
        fmt.Println(err.Error())
    }
    // Output:
    // error 1_Tabs: the given function did not return the expected HTML result.
    // given markdown: "\tfoo\tbaz\t\tbim\n"
    // expect HTML: "<pre><code>foo\tbaz\t\tbim\n</code></pre>\n"
    // actual HTML: "<p>Hello, World!</p>"
}
```

- [View it online](https://go.dev/play/p/cvzhbhEx_QG) @ Go Playground
- Supported CommonMark spec versions:
  - CommonMark [v0.13](https://spec.commonmark.org/0.13/) to [v0.30](https://spec.commonmark.org/0.30/) ([latest](https://spec.commonmark.org/current/))
- References on CommonMark:
  - [Markdown Reference](https://commonmark.org/help/) @ commonmark.org
  - [CommonMark specs](https://spec.commonmark.org/) @ spec.commonmark.org

## Contributing

[![go1.18+](https://img.shields.io/badge/Go-1.18+-blue?logo=go)](https://github.com/KEINOS/go-md-spec-check/blob/main/.github/workflows/unit-tests.yml#L81 "Supported versions")
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-md-spec-check.svg)](https://pkg.go.dev/github.com/KEINOS/go-md-spec-check/ "View document")
[![Opened Issues](https://img.shields.io/github/issues/KEINOS/go-md-spec-check?color=lightblue&logo=github)](https://github.com/KEINOS/go-md-spec-check/issues "opened issues")
[![PR](https://img.shields.io/github/issues-pr/KEINOS/go-md-spec-check?color=lightblue&logo=github)](https://github.com/KEINOS/go-md-spec-check/pulls "Pull Requests")

We are open to anything that helps us improve. We have 100% test coverage, so feel free to play with the code!

- Branch to PR: `main`
- Report an issue: [issues](https://github.com/KEINOS/go-md-spec-check/issues) @ GitHub
  - Please attach reproducible test cases. It helps us a lot.

## Statuses

[![UnitTests](https://github.com/KEINOS/go-md-spec-check/actions/workflows/unit-tests.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/unit-tests.yml)
[![PlatformTests](https://github.com/KEINOS/go-md-spec-check/actions/workflows/platform-tests.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/platform-tests.yml)
[![golangci-lint](https://github.com/KEINOS/go-md-spec-check/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/golangci-lint.yml)
[![CodeQL-Analysis](https://github.com/KEINOS/go-md-spec-check/actions/workflows/codeQL-analysis.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/codeQL-analysis.yml)

[![codecov](https://codecov.io/gh/KEINOS/go-md-spec-check/branch/main/graph/badge.svg?token=jW3haldEtr)](https://codecov.io/gh/KEINOS/go-md-spec-check)
[![Go Report Card](https://goreportcard.com/badge/github.com/KEINOS/go-md-spec-check)](https://goreportcard.com/report/github.com/KEINOS/go-md-spec-check)
[![Weekly Update](https://github.com/KEINOS/go-md-spec-check/actions/workflows/weekly-update.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/weekly-update.yml)

## License, copyright and credits

- MIT, Copyright (c) 2022 [KEINOS and the go-md-spec-check contributors](https://github.com/KEINOS/go-md-spec-check/graphs/contributors).
