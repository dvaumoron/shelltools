# shelltools

Some [CLI](https://en.wikipedia.org/wiki/Command-line_interface) utilities to combine within shell (inspired by [Nushell](https://www.nushell.sh/)).

shelltools use [expr-lang/expr](https://github.com/expr-lang/expr) (see [language definition](https://expr-lang.org/docs/language-definition)).

For more advanced SQL related need, look at [trdsql](https://github.com/noborus/trdsql).

Available utilities:

1. linetojson
2. jsonwhere
3. jsonorderby
4. jsontotable
5. cmdwithall
6. cmdforeach
7. distinctline

Examples of basic usage (all command are self documented with `--help`) :

<img src="https://raw.githubusercontent.com/dvaumoron/shelltools/main/screenshot/shelltools-screenshot.png">

## Getting started

Install via [Homebrew](https://brew.sh)

```console
$ brew tap dvaumoron/tap
$ brew install shelltools
```

Or get the [last binary](https://github.com/dvaumoron/shelltools/releases) depending on your OS.
