# phonenumbers Go tools

This directory contains the maintenance tools for the Go port of [libphonenumber](https://github.com/google/libphonenumber).

The library reads the master metadata under `resources/` at run time. The former Python metadata builders are therefore represented by the runtime loaders in `phonenumbers/buildmetadata.go` and `phonenumbers/prefix`; no generated region packages are required.

## Repository checks

The root `Makefile` provides the common development commands:

```console
make             # go build ./...
make test        # go test ./...
make vet         # go vet ./...
make fmt         # fail if any Go source needs gofmt
make cover       # run tests and write coverage.out
```

`go run ./tools/allcheck` checks that every non-generated Go package has exactly one package doc comment. Go exports identifiers by capitalization, so the Python `__all__` consistency check has no direct equivalent.

## Locale data

`tools/locale/DumpLocale.java` converts the installed JDK's `java.util.Locale` country names directly into `phonenumbers/geodata/locale_generated.go`. A JDK is the only prerequisite:

```console
make locale
```

The generated map keeps the first language spelling of a country name and represents duplicate spellings as `*<language>`, preserving the compact format used by the original data.

## Upstream release check

`tools/checkupstreamversion` queries the latest `google/libphonenumber` GitHub release and compares its tag with `phonenumbers.Version`:

```console
make upstream
```

The command exits with status 0 for a match, 1 for a version mismatch, and 2 for a network, response, or tag-format error. It uses only the Go standard library and builds without network access.

## Release checklist

1. Update `HISTORY.md` and `phonenumbers.Version`.
2. Run `make fmt`, `make vet`, `make test`, and `go run ./tools/allcheck`.
3. Optionally run `make locale` and review the JDK-derived data change.
4. Run `make upstream`, then create and push the `vX.Y.Z` tag.

The former `issue294.py` program was a one-off issue reproducer, not a maintenance tool, and has no Go replacement.
