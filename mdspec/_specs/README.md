# Test cases to be embedded

JSON files in this directory are the test cases from the [CommonMark spec](https://spec.commonmark.org/).

These `spec_*.json` files will be embedded in the binary and used to test the specification.

## How to update

To update:

1. See the [CommonMark spec](https://spec.commonmark.org/) for the latest test cases.
2. Edit the `spec_list.json` and run `download_specs.go` to download the latest test cases.
    - Note that the `spec_list.json` is ordered, from the most recent to the oldest version.
3. Update the test of `ExampleListVersion()` at "examples_test.go" to reflect the changes.
