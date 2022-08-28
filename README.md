# Markdown Specification Checker for Go

[![go1.16+](https://img.shields.io/badge/Go-1.16+-blue?logo=go)](https://github.com/KEINOS/go-md-spec-check/blob/main/.github/workflows/unit-tests.yml#L81 "Supported versions")
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-md-spec-check.svg)](https://pkg.go.dev/github.com/KEINOS/go-md-spec-check/ "View document")

[go-md-spec-check](https://github.com/KEINOS/go-md-spec-check) is a simple golang package that **checks if a function can convert Markdown to HTML according to the [CommonMark](https://commonmark.org/) specification**.

```go
// go 1.16+
go get github.com/KEINOS/go-md-spec-check
```

In the below example, `mdspec.SpecCheck()` runs `myMarkdownParser()` against about 500-600 test cases and checks that the results comply with the given version of the CommonMark specification.

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

- Supported CommonMark spec versions:
  - CommonMark [v0.13](https://spec.commonmark.org/0.13/) to [v0.30](https://spec.commonmark.org/0.30/) ([latest](https://spec.commonmark.org/current/))
- References on CommonMark:
  - [Markdown Reference](https://commonmark.org/help/) @ commonmark.org
  - [CommonMark specs](https://spec.commonmark.org/) @ spec.commonmark.org

## Statuses

[![UnitTests](https://github.com/KEINOS/go-md-spec-check/actions/workflows/unit-tests.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/unit-tests.yml)
[![PlatformTests](https://github.com/KEINOS/go-md-spec-check/actions/workflows/platform-tests.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/platform-tests.yml)
[![golangci-lint](https://github.com/KEINOS/go-md-spec-check/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/golangci-lint.yml)
[![CodeQL-Analysis](https://github.com/KEINOS/go-md-spec-check/actions/workflows/codeQL-analysis.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/codeQL-analysis.yml)

[![codecov](https://codecov.io/gh/KEINOS/go-md-spec-check/branch/main/graph/badge.svg?token=jW3haldEtr)](https://codecov.io/gh/KEINOS/go-md-spec-check)
[![Go Report Card](https://goreportcard.com/badge/github.com/KEINOS/go-md-spec-check)](https://goreportcard.com/report/github.com/KEINOS/go-md-spec-check)
[![Weekly Update](https://github.com/KEINOS/go-md-spec-check/actions/workflows/weekly-update.yml/badge.svg)](https://github.com/KEINOS/go-md-spec-check/actions/workflows/weekly-update.yml)

## Contributing

[![go1.16+](https://img.shields.io/badge/Go-1.16+-blue?logo=go)](https://github.com/KEINOS/go-md-spec-check/blob/main/.github/workflows/unit-tests.yml#L81 "Supported versions")
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-md-spec-check.svg)](https://pkg.go.dev/github.com/KEINOS/go-md-spec-check/ "View document")
[![Opened Issues](https://img.shields.io/github/issues/KEINOS/go-md-spec-check?color=lightblue&logo=github)](https://github.com/KEINOS/go-md-spec-check/issues "opened issues")
[![PR](https://img.shields.io/github/issues-pr/KEINOS/go-md-spec-check?color=lightblue&logo=github)](https://github.com/KEINOS/go-md-spec-check/pulls "Pull Requests")

We are open to anything that helps us improve.

- Branch to PR: `main`
- Report an issue: [issues](https://github.com/KEINOS/go-md-spec-check/issues) @ GitHub
  - Please attach reproducible test cases.

## License, copyright and credits

- MIT, Copyright (c) 2022 [KEINOS and the go-md-spec-check contributors](https://github.com/KEINOS/go-md-spec-check/graphs/contributors).
