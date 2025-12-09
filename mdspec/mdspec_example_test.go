package mdspec_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/KEINOS/go-md-spec-check/mdspec"
)

//nolint:revive // markdown in myMarkdownParser is not used but keeping it for example purposes.
func Example() {
	// Sample Markdown-to-HTML conversion function that does not do its job.
	myMarkdownParser := func(markdown string) (string, error) {
		// Do something with the given markdown and return the parsed HTML.
		parsedHTML := "<p>Hello, World!</p>"

		return parsedHTML, nil
	}

	// Check if the `myMarkdownParser()` complies with the CommonMark specification
	// version 0.30.
	err := mdspec.SpecCheck("v0.30", myMarkdownParser)
	if err != nil {
		if strings.Contains(err.Error(), "did not return the expected HTML result") {
			fmt.Println("The parser does not comply with the CommonMark specification.")
		}
	}
	// Output:
	// The parser does not comply with the CommonMark specification.
}

func ExampleListVersion() {
	list, err := mdspec.ListVersion()
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range list {
		fmt.Println(v)
	}
	// Output:
	// v0.31.2
	// v0.30
	// v0.29
	// v0.28
	// v0.27
	// v0.26
	// v0.25
	// v0.24
	// v0.23
	// v0.22
	// v0.21
	// v0.20
	// v0.19
	// v0.18
	// v0.17
	// v0.16
	// v0.15
	// v0.14
	// v0.13
}
