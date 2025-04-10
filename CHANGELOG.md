# Change Log (go-package-manager)

## 0.46.2

- **BREAKING CHANGE**: simple `gpm install` is now an alias for `go mod download`
- fix: `install` & `now` commands
- tests: add more tests

## 0.45.1

- refactor: update dependencies

## 0.44.0

- feat: add `clone` command
- fix: `password` command
- tests: add more tests

## 0.43.5

- ci: implement first tests
- fix: `base64` command and write tests for it
- fix: `list` command and write tests for it
- fix: `publish` command
- tests: add more tests

## 0.42.2

- **BREAKING CHANGE**: non-empty strings are now handled as `true` value for `CI` environment variable
- feat: `environment` section for `gpm.yaml` files

## 0.41.4

- **BREAKING CHANGE**: implement feature of settings in `settings.yaml` and `gpm.yaml` files and refactored many settings to it
- refactor: code cleanups and improvements

## 0.40.1

- **BREAKING CHANGE**: `publish` command now supports `publish`, `prepublish` and `postpublish` scripts in `gpm.yaml` file

## 0.39.0

- **BREAKING CHANGE**: `bump` command now supports `bump`, `prebump` and `postbump` scripts in `gpm.yaml` file
- **BREAKING CHANGE**: `start` command now supports `prestart` script in `gpm.yaml` file
- **BREAKING CHANGE**: `tidy` command now supports `pretidy` and `posttidy` scripts in `gpm.yaml` file
- refactor: `--temperature` flag is now globally
- refactor: `--no-pre-script`, `--no-post-script` and `--no-script` flags are now globally
- fix: `test` command did not make use of `pretest` and `posttest` scripts
- fix: more minor fixes

## 0.38.0

- **BREAKING CHANGE**: `generate powerpoint` command now uses regular expressions now

## 0.37.0

- **BREAKING CHANGE**: `generate powerpoint` command now uses [.gitignore-like glob patterns](https://github.com/sabhiram/go-gitignore)
- **BREAKING CHANGE**: `import aliases` and `import projects` commands now uses default sources if no one is defined in the arguments ... both defaults can be customized by `GPM_DEFAULT_ALIAS_SOURCE` and/or `GPM_DEFAULT_PROJECT_SOURCE` environment variables
- feat: add `--max-slides`, `--min-slides` and `--temperature` flags for `generate powerpoint` command
- refactor: can use custom standard input now and removed `LoadFromSTDINIfAvailable()` function for this
- refactor: code cleanups

## 0.36.0

- **BREAKING CHANGE**: `AppContext.CreateAIChat()` now uses initial value from `--system-prompt` flag
- feat: `generate powerpoint` command, which generates PowerPoint `.pptx` file from files using AI
- refactor: code cleanups and improvements

## 0.35.1

- **BREAKING CHANGE**: `generate guid` => `guid`/`uuid`
- **BREAKING CHANGE**: `generate password` => `password`
- feat: `base64` command also supports input files and `--data-uri` flag now
- feat: `compress` command which compresses data with GZIP
- feat: `cron` command that runs any kind of "executables" periodically
- feat: `sleep` command that waits for a specific number of time
- feat: `uncompress` command which uncompresses GZIP data

## 0.34.3

- feat: `base64` command that outputs binary input data base Base64 string
- feat: `cat` command that outputs binary input data combined to STDOUT
- feat: `describe` command that describes images
- feat: `now` command that outputs current time
- refactor: `--model` flag is now global

## 0.33.0

- feat: enhance features of `gpm.yaml` file

## 0.32.0

- **BREAKING CHANGE**: when an environment is defined, the root base path of the application changes from `${HOME}/.gpm` to the specified path
- **BREAKING CHANGE**: with the introduction of the `GPM_ROOT_BASE_PATH` environment variable, `.env.{ENV-SUFFIX}` files are no longer supported ... these files must be moved or merged into the corresponding environment subfolder, like `${HOME}/.gpm/<ENV-NAME>`, as `.env`
- **BREAKING CHANGE**: Scripts in the `gpm.yaml` file with an environment prefix (e.g., `dev:build` for a `dev` environment) will now execute instead of their unprefixed counterparts (e.g., `build`)
- feat: added support for the `GPM_ROOT_BASE_PATH` environment variable
- feat: you can now define a custom `files` section with an environment suffix in the `gpm.yaml` file (e.g., `files:dev` for a `dev` environment)
- feat: can define custom path to `aliases.yaml` file by `GPM_ALIASES_FILE` environment variable
- feat: can define custom path to `projects.yaml` file by `GPM_PROJECTS_FILE` environment variable

## 0.31.0

- feat: `generate password` command
- feat: `generate uuid` command

## 0.30.0

- feat: `gpm update` now supports update of specific modules as well
- fix: using `TidyUp()` method from `AppContext` instead raw `go mod tidy`
- refactor: `gpm doctor` now checks all dependencies for up-to-dateness
- refactor: code cleanups and improvements

## 0.29.10

- feat: self-update by executing `gpm update --self`
- chore: improve (self-)update scripts
- fix: `doctor` command

## 0.28.0

- **BREAKING CHANGE**: `bump version` command is now reduced to simple `bump`
- feat: additional checks like environment variables and unsed modules by `doctor` command

## 0.27.0

- **BREAKING CHANGE**: remove `audit` command
- feat: `doctor` command

## 0.26.2

- feat: automatic Go compiler detection in `generate project` command
- refactor: using Go compiler `1.23.4`
- fix: `generate project` command

## 0.25.5

- feat: `publish` command
- feat: add build for `windows/arm64`
- fix: `push` command

## 0.24.0

- feat: `update` command which updates all dependencies of current project
- refactor: updated all dependencies

## 0.23.0

- refactor: `generate project` now runs in a terminal UI
- fix: `show dependencies` that can now handle big diagram sources as well

## 0.22.5

- **BREAKING CHANGE**: default AI model for Open AI is `gpt-4o-mini` now
- **BREAKING CHANGE**: default AI model for Ollama is `llama3.3` now
- feat: `generate project` command that generates a Go project with natural language
- docs: can directly update `gpm` from `sh.kloubert.dev` server now

## 0.21.3

- ci: automatic update of [GitHub wiki](https://github.com/mkloubert/go-package-manager/wiki) as well

## 0.20.30

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
