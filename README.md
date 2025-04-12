# Zephyranthes Archiver & Backups

Zephyranthes is a tool for creating archives, intended for mainly for backing up
files. A description of the archives to create is provided as a sequence of
"backup specs" to standard input, or as a file argument containing the
definitions. Multiple files may be provided as arguments. For each backup spec,
an archive is generated.

## Usage

```sh
zephyr -h
usage: zephyr [options] [file ...]

Options:

  -dry-run
    
  -h    Show usage.
  -help
        Show usage.
  -log-file string
        Log what we're doing to the specified FILE.
  -log-level value
        How verbose the log file is. One of: fatal, error, warning, info, verbose, debug
  -v    Produce verbose output.
  -verbose
        Produce verbose output.

Each file is parsed to define the backup archive(s) to create. Defaults to reading from standard input.
```

## Backup Specs

Backups may be defined in either YAML or JSON format. The file is a list of
backup specifications that define the archives to be created, and what its
contents are. Each backup specification defines one archive, and a file may
define one or more backup specifications.

The following YAML defines a zip archive named "backup.zip" to be created in the
system's root directory, containing the contents of two system directories.

```yaml
- name: Name of the backup
  path: /backup.zip
  format: zip
  contents:
    - /etc
    - /usr/local/etc
```

The following JSON defines the same thing, except as an uncompressed tape archive.

```json
[
    "backups": [
        {
            "name": "Name of the backup",
            "path": "/backup.tar",
            "format": "tar",
            "contents": [
                "/etc",
                "/usr/local/etc"
            ]
        }
    ]
]
```

### Formats

The `format` field can be one of the specified values:

| Value     | Output format       |
| --------- | ------------------- |
| "zip"     | Zip archive         |
| "tar"     | TAR archive         |
| "tgz"     | Gzip compressed TAR |
| "tar.gz"  | Alias for tgz       |