# Go package documentation

The Python `pydoc` HTML pages have been superseded by Go's built-in documentation. From the module root, use:

```console
go doc ./phonenumbers
go doc -all ./phonenumbers
go doc ./phonenumbers.Parse
```

The module contains these packages:

- `phonenumbers`: parsing, validation, formatting, matching, as-you-type formatting, and short-number information.
- `phonenumbers/geocoder`: localized geographic descriptions.
- `phonenumbers/carrier`: original carrier-name lookup.
- `phonenumbers/timezone`: time-zone lookup.
- `phonenumbers/prefix`: loaders and longest-prefix lookup used by the extended packages.
- `phonenumbers/geodata`: generated locale display names used by the geocoder.
- `resources`: embedded libphonenumber source metadata.
- `internal/pyre` and `internal/pyrt`: compatibility support for CPython regular-expression and Unicode behavior; these cannot be imported outside this module.
- `cmd/phonenumbers` and `cmd/appspot`: command-line compatibility evaluators.
- `tools/allcheck` and `tools/checkupstreamversion`: repository maintenance commands.

For a package outside the core library, pass its full relative path, for example `go doc ./phonenumbers/geocoder`.
