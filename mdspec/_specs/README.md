# Test cases to be embedded

JSON files in this directory are the test cases from the [CommonMark spec](https://spec.commonmark.org/).

These `spec_*.json` files will be embedded in the binary and used to test the specification.

## How to update

1. Move to `_updater` directory in the parent directory.
2. Run the `download_specs.go` program to download the latest test cases.
