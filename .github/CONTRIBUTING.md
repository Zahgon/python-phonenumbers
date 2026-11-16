# Contributing to python-phonenumbers

Contributions are welcome, but please be aware that the project is directly
derived from the
[upstream Java project](https://github.com/google/libphonenumber).

As a result, any changes to the phonenumber metadata or to the API should be
[submitted there](https://github.com/google/libphonenumber/pulls)
rather than here; only issues or contributions that are specific to the Go
version are likely to be relevant here.

Format changes with `gofmt`, then run `go test ./...` and `go vet ./...` before
submitting a pull request. The module supports Go 1.23 and later, so new code
must not depend on APIs introduced after Go 1.23.
