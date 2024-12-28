# Change Log (go-package-manager)

## 0.20.19

- docs: will now be build for [gpm.kloubert.dev](https://gpm.kloubert.dev)
- feat: `push` command now supports `--with-tags` flag to additionally push tags after upload changes to remote(s)
- fix: `run` command
- fix: `GetBestChromaFormatterName()` function

## 0.19.0

- feat: `chat` command is available on OpenBSD again

## 0.18.4

- **BREAKING CHANGE**: using `checkout` command with `:` prefix in alias name, will handle this name as alias for a final target branch ... this can be defined via environment variables in the format `GPM_BRANCH_<ALIAS-NAME>` now
- **BREAKING CHANGE**: `execute` command now generates shell commands from natural language using AI
- feat: `generate documentation` command, that creates documentation files of this tool
- fix: `down` command
- fix: build fixes for different systems and architechtures

## 0.17.0

- feat: `show dependencies` command, which generates and opens HTML page with a dependency graph of the modules of the current project

## 0.16.0

- feat: `down` and `up` commands, which are shorthands for `docker compose down` and `docker compose commands`
- feat: `monitor` command now also displays open network and files information
- feat: `prompt` command which invokes an AI prompt or (continues) a chat from command line
- feat: `setup updater` now supports custom target folder for binary with `--install-path` flag and `GPM_INSTALL_PATH` environment variable
- fix: `AppContext.RunShellCommandByArgs()` uses `AppContext.Cwd`
- refactor: code cleanups and improvements

## 0.15.0

- **BREAKING CHANGE**: `pack` command creates corresponding `.sha256` file by default
- feat: `monitor` command, which displays CPU and memory usage of a process in the console
- feat: `setup git` command, which sets up basic git features like username and email address
- feat: support for `CI` environment variable, which is used by GitHub actions or GitLab runners as well
- refactor: using new and shorter `CheckForError()` function instead of `CloseWithError()` in most cases
- refactor: improve logging
- refactor: using Go template engine now to generate final content of script by `setup updater` command
- chore: update `projects.yaml` with: [hugo](https://github.com/gohugoio/hugo)

## 0.14.0

- **BREAKING CHANGE**: `audit` command now uses spinners and pretty tables for output
- **BREAKING CHANGE**: if `GPM_BIN_PATH` is relative, it will be mapped to `$HOME/.gpm` instead
- feat: `bump version` command, which upgrades the current version of the underlying repository by setting up a new Git tag locally
- feat: `diff` command, which displays changes between version (or the current HEAD) as pretty diff output
- feat: `init` command, which currently can initialize a `gpm.yaml` file
- feat: add `GPM_AI_CHAT_TEMPERATURE` environment variable, which defines the custom temperature value for AI chat (operations)
- feat: add `--temperature` flag to `chat` command, which can define the initial temperature value for the command
- feat: `setup updater` command, which installs a shell script called `gpm-update` in `$HOME/.gpm/bin` folder of a UNIX environment, like \*BSD, Linux or MacOS
- refactor: improve prompting in `chat` command
- refactor: `pack` command now outputs progress with pretty bars
- refactor: code cleanups and improvements

## 0.13.0

- feat: `chat` command, which is a simple user interface for communicating with AI chat bots via [Ollama](https://ollama.com) or [ChatGPT / OpenAI](https://platform.openai.com/docs/api-reference)
- feat: `audit` command, which uses [osv.dev API](https://google.github.io/osv.dev/api/) to scan installed modules for vulnerabilities
- feat: `install` command now supports `--tidy` flag to run `tidy` command after successful installation
- feat: `open alias` command, which opens the URL of an alias from `aliases.yaml` file in `$HOME/.gpm/bin` folder
- feat: `open project` command, which opens the URL of a project from `projects.yaml` file in `$HOME/.gpm/bin` folder
- feat: `import aliases` command, which loads aliases from a local or web source and merge them with `aliases.yaml` file in `$HOME/.gpm` folder
- feat: `import projects` command, which loads projects from a local or web source and merge them with `projects.yaml` file in `$HOME/.gpm` folder
- feat: implement `postinstall`, `postpack` and `prepack` scripts for `gpm.yaml` files

## 0.12.0

- feat: `make` command, which downloads a Git repository, then runs `build` command from it and move the final executable to `$HOME/.gpm/bin` folder ... command is also able to handle aliases create by `add alias` command as well
- feat: can setup `GPM_BIN_PATH` environment variable for a custom central folder for binaries, which is `$HOME/.gpm/bin` by default
- feat: `remove binary` command, which removes binary installed with `make` command
- feat: `list binaries` command, which lists binaries installed with `make` command
- feat: `pack` command, which compresses files defined in `files` section of `gpm.yaml` file to zip archive
- fix: setup `Dir` property of commands used in Git\* methods of `AppContext` instance
- chore: code cleanups and improvements

## 0.11.0

- feat: load `.env` files from `$HOME/.gpm` and project directories, if exist, automatically
- feat: add `--env-file` file to load environment variables from external variables
- feat: `execute` command which runs shell commands with the environment variables loaded from .env\* files
- fix: exit app if special files could not be loaded
- feat: `run` command without scripts will run `go run .` now
- feat: `postbuild`, `postinstall`, `posttest`, `prebuild`, `preinstall` and `pretest` script support
- feat: add `--no-script` flags for `build`, `start`, `test` and `tidy`

## 0.10.1

- initial release
