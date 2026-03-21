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
- **--dry-run**: Don't actually create an archive.
- **-h**, **--help**: Show usage summary.
- **--log-file FILE**: Log what we're doing to the specified **FILE**.
- **--log-level LEVEL**: How verbose the log file is.
  - One of: **fatal**, **error**, **warning**, **info**, **verbose**, **debug**.
- **--man**: Show this manual page.
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
- **contents**: list of files to backup in the archive.

These "backup spec" objects defines a single archive, and each backup
specification file defines one or more backup spec. There is no practical limit
on the number of backup of archives that may be specified in one file.

Fields that define a file path are subject to environment variable expansion
using syntax similar to Bourne shells. "`$HOME`" and "`${HOME}`" may both be
used to expand the environment variable "`HOME`".

Relative paths are resolved using the current working directory. This may be
controlled naturally by changing the startup directory of the program before
running zephyr, or by using the **-C** or **--directory** option to change the
working directory before any backups are run.

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
    - path: "/mnt/${USER}.zip"
    - format: zip
    - contents:
        - Documents
        - Pictures
        - Videos
    - name: Backup system files.
    - path: "/mnt/${HOSTNAME}.tgz"
    - format: tgz
    - contents:
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
