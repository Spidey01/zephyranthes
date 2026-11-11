# ZEPHYR(1) -- Create archives from backup specifications

## SYNOPSIS

`zephyr [options] file ...`

## DESCRIPTION

Zephyranthes is a tool for creating archives, intended mainly for backing up
files. The executable is called **zephyr** and takes zero or more files
specifying the backup archives to be created.  If no input files are specified,
Zephyranthes reads backup specs from standard input.  For each backup spec
defined, an archive is generated.

## OPTIONS

- **-C DIR**, **--directory DIR**: Change directory before opening and running
  the backup specs.
- **--about**: Show copyright and license information about the program.
- **--dry-run**: Don't actually create an archive.
- **-h**, **--help**: Show usage summary.
- **--log-file FILE**: Log what we're doing to the specified **FILE**.
- **--log-level LEVEL**: How verbose the log file is.
  - One of: **fatal**, **error**, **warning**, **info**, **verbose**, **debug**.
- **--man**: Show this manual page.
- **--no-exec**: Do not execute external commands defined in the backup spec.
- **-v**, **--verbose**: Produce verbose output.
- **--version**: Show version info and exit.

## BACKUP SPECS

Backups may be defined in either YAML or JSON format. The file is a list of
backup specifications that define the archives to be created, and what its
contents are. Each backup specification defines one archive, and a file may
define one or more backup specifications as a list of objects.

Each backup specification object contains the following fields:

- **name**: name of the backup. This may be used in logging, command output, or
  even incorporated into archvie metadata as a comment.
- **path**: path of the output archive.
- **format**: format of the output archive. One of the following values:
  - **zip**: archive will be formatted as ZIP.
  - **tar**: plain tape archive (`.tar`) file.
  - **tgz**, **tar.gz**: tape archive compressed with [**gzip**(1)](man(gzip))
  - **tzst**, **tar.zst**: tape archive compressed with [**zstd**(1)](man(zstd))
- **contents**: list of files to backup in the archive.
- **before**, **after**: external commands to run before/after contents are added to the archive.
  - **commands**: list of commands to execute. Does not invoke the shell.
  - **directory**: overrides starting directory of commands.
  - **environment**: overrides starting environment of commands.
  - **timeout**: limits runtime of the commands list; parsed as a duration string.

These "backup spec" objects defines a single archive, and each backup
specification file defines one or more backup spec. There is no practical limit
on the number of backup of archives that may be specified in one file. Backups
are executed sequentually in the order defined.

Each item in the `contents` list supports file globbing. E.g.,
`/etc/foo.d/*.conf` will archive the matching files.

Fields that define a file path are subject to environment variable expansion
using syntax similar to Bourne shells. "`$HOME`" and "`${HOME}`" may both be
used to expand the environment variable "`HOME`".

Relative paths are resolved using the current working directory. This may be
controlled naturally by changing the startup directory of the program before
running zephyr, or by using the **-C** or **--directory** option to change the
working directory before any backups are run.

Executing `before` occurs before the archive is created. If any of the commands
list returns non-zero, the backup fails and no archive is created. This can be
used to prepare for the backup. For example dumping a database to a path
specified in `contents` or mounting the filesystem where `path` will be written.

Executing `after` occurs after the archive has been closed. If the archive
failed the `after` commands are executed to ensure they can cleanup actions
taken by `before` commands.  If any of the `after` commands list returns
non-zero, the backup is considered a failure.

Executing external commands is enabled by default and can be disabled using the
`--no-exec` option. When `--dry-run` is set, commands will be printed to
standard output but not executed. Commands inherit the current working directory
and environment; to unset some environment variable "foo" define it as an empty
string in the override list. Commands are specified as a list. Each element is
expanded for environment variables including those defined in the overrides,
split by spaces to determine the argument vector, and then executed. Command
output will only inherit the standard output and error streams when the `-v` or
`--verbose` options are set. Commands are executed directly, if redirection or
pipes are required then its necessary to execute a script instead. The timeout
duration applies to the entire command list.

A duration string is a possibly signed sequence of decimal numbers, each with
optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid
time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

## EXAMPLES

- **Run a backup**

    ```bash
    zephyr backup.json
    ```

- **Run a set of backup files, saving log output to `log.txt`**

    ```bash
    zephyr --log-level info --log-file log.txt sysbkup.yml usrbkup.yml
    ```

- **Contents of a backup spec in JSON.** In this example, multiple archives are created.

    ```json
    [
        {
            "name": "Backup of user files.",
            "path": "/mnt/${USER}.zip",
            "format": "zip",
            "contents": [
                "Documents",
                "Pictures",
                "Videos"
            ]
        },
        {
            "name": "Backup of system files.",
            "path": "/mnt/${HOSTNAME}.tgz",
            "format": "tgz",
            "contents": [
                "/etc",
                "/usr/local"
            ]
        }
    ]
    ```

- **Contents of a backup spec in YAML.** In this example, multiple archives are created.

    ```yaml
    - name: Backup user files.
      path: "/mnt/${USER}.zip"
      format: zip
      contents:
        - Documents
        - Pictures
        - Videos
    - name: Backup system files.
      path: "/mnt/${HOSTNAME}.tgz"
      format: tgz
      contents:
        - /etc
        - /usr/local
    ```

- **Various ways to view this manual**

    ```bash
    zephyr --man | less
    zephyr --man | glow -s
    zephyr --man | pandoc -s -t man | groff --mandoc -Tascii | more
    ```

## AUTHOR

Terry Poulin

## BUGS

Please report via the [Zephyranthes repo on GitHub](https://github.com/Spidey01/zephyranthes).

## SEE ALSO

[**tar**(1)](tar(1)), [**unzip**(1)](unzip(1)), [**pandoc**(1)](https://pandoc.org/MANUAL.html), [**glow**(1)](https://github.com/charmbracelet/glow)
