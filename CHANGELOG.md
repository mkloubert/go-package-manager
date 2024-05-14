# Change Log (go-package-manager)

## 0.11.0

- feat: load `.env` files from `$HOME/.gpm` and project directories, if exist, automatically
- feat: add `--env-file` file to load environment variables from external variables
- feat: `execute` command which runs shell commands with the environment variables loaded from .env* files
- fix: exit app if special files could not be loaded
- feat: `run` command without scripts will run `go run .` now
- feat: `build` command now supports `--no-script` flag

## 0.10.1

- initial release
