# benchfmt

Formats go benchmarking into ASCII or MD

## Installation

```shell
go get github.com/sirkon/benchfmt@latest
```

```shell
mise use --global go:github.com/sirkon/benchfmt@latest
```

## Usage

```shell
go test -bench=. | benchfmt    # For ASCII output, very close to prettybench output.
go test -bench=. | benchfmt md # For Markdown output
```
