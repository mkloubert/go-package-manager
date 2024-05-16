# Change Log (go-package-manager)

## 0.13.0

- feat: `install` command now supports `--tidy` flag to run `tidy` command after successful installation
- feat: `open alias` command, which opens the URL of an alias from `alias.yaml` file in `$HOME/.gpm/bin` folder
- feat: `open project` command, which opens the URL of a project  from `projects.yaml` file in `$HOME/.gpm/bin` folder

## 0.12.0

- feat: `make` command, which downloads a Git repository, then runs `build` command from it and move the final executable to `$HOME/.gpm/bin` folder ... command is also able to handle aliases create by `add alias` command as well
- feat: can setup `GPM_BIN_PATH` environment variable for a custom central folder for binaries, which is `$HOME/.gpm/bin` by default
- feat: `remove binary` command, which removes binary installed with `make` command
- feat: `list binaries` command, which lists binaries installed with `make` command
- feat: `pack` command, which compresses files defined in `files` section of `gpm.yaml` file to zip archive
- fix: setup `Dir` property of commands used in Git* methods of `AppContext` instance
- chore: code cleanups and improvements

## 0.11.0

- feat: load `.env` files from `$HOME/.gpm` and project directories, if exist, automatically
- feat: add `--env-file` file to load environment variables from external variables
- feat: `execute` command which runs shell commands with the environment variables loaded from .env* files
- fix: exit app if special files could not be loaded
- feat: `run` command without scripts will run `go run .` now
- feat: `postbuild`, `postinstall`, `posttest`, `prebuild`, `preinstall` and `pretest` script support
- feat: add `--no-script` flags for `build`, `start`, `test` and `tidy` 

## 0.10.1

- initial release
