# Markdown Specification Checker for Go

`github.com/KEINOS/go-md-spec-check` is a simple golang package that checks if a function can convert Markdown to HTML according to the [CommonMark](https://commonmark.org/) specification.

```go
// go 1.16+
go get github.com/KEINOS/go-md-spec-check
```

In the below example, `mdspec.SpecCheck()` runs `myMarkdownParser()` against about 500 test cases and checks that the results comply with the given version of the CommonMark specification.

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

## References

- [Markdown Reference](https://commonmark.org/help/) @ commonmark.org
- [CommonMark specs](https://spec.commonmark.org/) @ spec.commonmark.org
