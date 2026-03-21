# Zephyranthes Archiver & Backups

Zephyranthes is a tool for creating archives, intended for mainly for backing up
files. A description of the archives to create is provided as a sequence of
"backup specs" to standard input, or as a file argument containing the
definitions. Multiple files may be provided as arguments. For each backup spec,
an archive is generated.

## Usage

```sh
zephyr --help
usage: zephyr [options] [file ...]

Options:

  -C string
        Alias for --directory.
  -directory string
        Change directory before opening and running the backup specs.
  -dry-run
    
  -h    Show usage.
  -help
        Show usage.
  -log-file string
        Log what we're doing to the specified FILE.
  -log-level value
        How verbose the log file is. One of: fatal, error, warning, info, verbose, debug
  -man
        Show manual page
  -v    Produce verbose output.
  -verbose
        Produce verbose output.
  -version
        Show version info and exit.

Each file is parsed to define the backup archive(s) to create. Defaults to reading from standard input.
```

For more information, see the [man page](zephyr.1.md).

## Archive Formats

Zephyranthes can create archives in the specified formats.

| Value     | Output format       |
| --------- | ------------------- |
| "zip"     | Zip archive         |
| "tar"     | TAR archive         |
| "tgz"     | Gzip compressed TAR |
| "tar.gz"  | Alias for tgz       |

## Installation

For a system that has `GOBIN` in `PATH`, simply run

```sh
go install
```

The executable can also be compiled and then copied into `PATH` as you see fit.

```sh
go build -o zephyr
```

If you need to cross-compile for another operating system or machine
architecture, set the `GOOS` and `GOARCH` environment variables before
compiling. See `go tool dist list` for a list of combinations.
