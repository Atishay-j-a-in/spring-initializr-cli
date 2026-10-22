# Implementation Plan: Dependency-Free Spring Initializr CLI in Go

## 1. Goal

Build a small, interactive terminal application that creates Spring Boot
projects by calling the official Spring Initializr service
(`https://start.spring.io`).

The project is intended for an AI coding agent to implement step by
step. Follow this document as the source of truth. Do not add features
or dependencies that contradict it.

**Working project name:** `spring-create`\
**Executable name:** `spring-create` (`spring-create.exe` on Windows)\
**Initial development OS:** Windows 10/11, 64-bit\
**Implementation language:** Go\
**Third-party Go modules:** None. Use only the Go standard library.\
**Runtime services:** Internet access to `start.spring.io` is required
to load current metadata and download a generated project.

This is a terminal CLI, not a GUI application. The first release should
favor reliability and clear numbered/selectable prompts over a complex
full-screen terminal UI. It must work in Windows Terminal and PowerShell
without requiring Java to run the CLI.

## 2. Product requirements

### 2.1 Main user journey

Running `spring-create` with no arguments starts an interactive wizard:

1.  Check that standard input and output are usable. If input is
    redirected, print a useful error and exit nonzero rather than
    hanging.
2.  Fetch Spring Initializr metadata from `https://start.spring.io` with
    a timeout.
3.  Display a short welcome message and the service URL.
4.  Ask for project settings using defaults derived from metadata
    whenever possible:
    -   Project name / artifact ID
    -   Group ID
    -   Artifact ID
    -   Base package name
    -   Description
    -   Language (Java, Kotlin, or Groovy if supported by metadata)
    -   Build system (Maven or Gradle variants supported by metadata)
    -   Packaging (jar/war if supported)
    -   Java version (only values supported by current metadata)
    -   Spring Boot version (only values supported by current metadata)
    -   Dependencies (multi-select from the service's current dependency
        metadata)
    -   Output directory
5.  Show a complete summary of the configuration.
6.  Ask for explicit confirmation before making the network request to
    generate the project.
7.  Request the ZIP from Spring Initializr.
8.  Validate the HTTP response and ZIP contents.
9.  Extract the project to the selected output directory, safely.
10. Print the resulting absolute path and next-step commands appropriate
    to the chosen build system.

### 2.2 Interaction design

-   Use Go standard library only.
-   Use simple line-based prompts and numbered menus. Do **not** use
    external TUI libraries, terminal raw-mode packages, or
    ANSI-dependent full-screen rendering in version 1.
-   Every prompt must show its default, where applicable.
-   Pressing Enter accepts a displayed default.
-   Menus must show numbered choices. For multi-select, accept
    comma-separated numbers, for example `1,3,5`; blank input means no
    optional selections.
-   Trim whitespace. Reject invalid input with a short explanation and
    reprompt.
-   Support `q` or `quit` at every wizard prompt to cancel cleanly.
-   Support Ctrl+C without stack traces. If the process is interrupted,
    exit cleanly and do not leave a partial generated project behind.
-   Keep output readable in plain Windows Terminal, PowerShell, cmd.exe,
    and redirected logs. Color is optional and must never be required to
    understand output.
-   Do not depend on the `TERM` environment variable.

### 2.3 Metadata-driven behavior

Do not hardcode a stale list of Java versions, Spring Boot versions,
build systems, languages, packaging types, or dependencies. Fetch
current metadata from Spring Initializr.

-   Request `GET https://start.spring.io/` with
    `Accept: application/vnd.initializr.v2.3+json`.
-   Set a descriptive `User-Agent`, such as `spring-create/0.1.0`.
-   Parse the response using `encoding/json`.
-   Model only the metadata fields needed by the CLI, but preserve
    unknown JSON fields where useful so future schema changes do not
    unnecessarily break parsing.
-   Inspect a real response from the service during implementation. Do
    not assume the metadata JSON has a particular shape without
    verifying it.
-   Use metadata defaults and available values to populate menus.
-   Dependency choices must come from metadata. Show each dependency's
    display name and ID; send dependency IDs in the generation request.
-   If the metadata request fails, show the URL, HTTP/network error, and
    a practical suggestion to check connectivity. Do not silently fall
    back to hardcoded versions.
-   If the metadata schema is missing a required field or changes
    incompatibly, return a clear error with enough context for
    debugging.

### 2.4 Generation request

-   Generate the project through the official Spring Initializr endpoint
    `GET https://start.spring.io/starter.zip` using query parameters
    supported by the current service metadata.
-   Use `net/url` to encode query parameters; never concatenate
    user-provided values directly into a URL.
-   At minimum, send the selected values for type/build, language,
    bootVersion, baseDir, groupId, artifactId, name, description,
    packageName, packaging, javaVersion, and dependencies, using the
    actual parameter names supported by the service.
-   Do not send empty optional values unless the service expects them.
-   Encode multiple dependencies in the exact format accepted by Spring
    Initializr; verify this against the official API and a real request.
-   Set the same `Accept` and `User-Agent` headers as appropriate.
-   Use a configurable timeout (default 60 seconds) and cap the maximum
    downloaded ZIP size (default 100 MiB; make this a named constant).
-   Require a successful HTTP status. For non-2xx responses, read only a
    bounded amount of the response body and display the HTTP status plus
    a readable error. Do not print unlimited server responses.
-   Do not log secrets or dump large ZIP content to the terminal.

### 2.5 ZIP extraction and filesystem safety

The generated ZIP is remote input. Treat it as untrusted.

-   Use `archive/zip`, `io`, `os`, `path/filepath`, and `strings` from
    the standard library.
-   Reject absolute paths, parent traversal (`..`), paths that resolve
    outside the intended output directory, and malformed ZIP entries.
-   Normalize ZIP entry paths safely before joining them to the
    destination.
-   Create directories with reasonable permissions and files without
    following any path outside the destination.
-   Do not overwrite existing user files silently.
-   If the target directory already exists and is non-empty, stop and
    explain the conflict unless the user explicitly chooses a new
    directory. Version 1 should not implement a force-overwrite flag.
-   Extract into a temporary staging directory first. Only move/rename
    the completed project into its final location after extraction
    succeeds. Clean up staging data on failure or cancellation.
-   Check for duplicate or conflicting ZIP paths where practical.
-   Enforce the downloaded ZIP size cap and a reasonable total
    uncompressed-size cap (default 250 MiB) to reduce ZIP-bomb risk.
-   Close all files and HTTP response bodies; propagate errors instead
    of ignoring them.
-   Never execute generated files or automatically run Maven/Gradle.

### 2.6 Error handling and exit codes

Use consistent error messages and exit codes:

-   `0`: success, help/version output, or user cancellation.
-   `1`: unexpected internal error.
-   `2`: invalid command-line arguments or invalid user input after
    retry limit (prefer reprompting for interactive input).
-   `3`: metadata/API/network failure.
-   `4`: ZIP validation, extraction, or filesystem failure.

Errors must go to stderr. Normal prompts and success messages go to
stdout. Avoid stack traces in normal operation; optionally show
additional diagnostics only when a future `--verbose` option is
implemented.

### 2.7 Command-line flags

Implement these flags in addition to the no-argument wizard:

-   `--help` / `-h`: print usage.
-   `--version`: print the compiled application version.
-   `--url`: override the Initializr base URL, for testing or custom
    Initializr instances. Validate it as an absolute `http` or `https`
    URL. Default: `https://start.spring.io`.
-   `--output DIR`: set the output parent directory; in wizard mode, use
    it as the default.
-   `--no-color`: accepted for forward compatibility; output must
    already be readable without color.
-   `--verbose`: print useful diagnostic context to stderr, without
    dumping ZIP data.

Version 1 does not need a fully non-interactive configuration mode. If
arguments are provided that require values, validate them and show usage
on errors.

### 2.8 Security and privacy

-   Never invoke a shell to build URLs or extract files.
-   Never execute commands from the generated project.
-   Never send user settings anywhere except the selected Initializr
    service.
-   Do not collect telemetry.
-   Do not store project settings unless a later feature explicitly adds
    a local config file.
-   Use HTTPS by default. Permit HTTP only when a user explicitly passes
    a custom `--url` (mainly for local testing).
-   Apply timeouts to all network operations.
-   Validate project coordinates and package names. Do not silently
    rewrite values in surprising ways; explain invalid input and request
    correction.
-   Use standard-library TLS verification; do not disable certificate
    validation.

## 3. Installation requirements for the developer

The developer currently has no Go tooling installed and uses Windows.
Install only the items below.

### Required

1.  **Go SDK (compiler and tools)**
    -   Open the official Go download page: https://go.dev/dl/

    -   Download the Windows installer (`.msi`) for **Windows x86-64 /
        amd64** unless the computer is ARM-based.

    -   Run the installer using its defaults.

    -   Close and reopen PowerShell or Windows Terminal after
        installation.

    -   Verify:

        ``` powershell
        go version
        go env GOROOT GOPATH
        ```

    -   `go version` must print a Go version. Use the latest stable Go
        release available when installing.
2.  **A code editor**
    -   VS Code is recommended but not mandatory. Any editor that can
        edit `.go` and `.md` files is acceptable.
    -   If using VS Code, the Go extension is optional; it improves
        editing but is not needed to compile.
3.  **Git (optional)**
    -   Git is needed only if the project is cloned from a Git
        repository or version control is desired. It is not required to
        compile a local folder.
    -   Official downloads: https://git-scm.com/download/win

### Not required

-   Java or a JDK to run this Go CLI.
-   Maven or Gradle to build the CLI.
-   Node.js/npm.
-   Python.
-   C/C++ compiler or CGO.
-   Docker.
-   Any Go third-party module or package.
-   A terminal framework.

The generated Spring Boot project will, of course, need the relevant
Java JDK and Maven/Gradle when the user later wants to build or run that
generated project. Those are requirements of the generated project, not
of this CLI.

### Optional editor setup

If VS Code is used: 1. Install VS Code from
https://code.visualstudio.com/. 2. Open the project folder. 3. When
prompted, install the official Go extension. 4. Do not add external Go
libraries merely to satisfy editor suggestions.

## 4. Create the project

Open PowerShell in the directory where the project should live:

``` powershell
mkdir spring-create
cd spring-create
go mod init example.com/spring-create
```

For a personal/local project, `example.com/spring-create` is a valid
placeholder module path. If the project is published, replace it with
the real repository module path before release.

Create this initial layout:

``` text
spring-create/
├── go.mod
├── README.md
├── LICENSE
├── cmd/
│   └── spring-create/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── cli/
│   │   ├── args.go
│   │   ├── prompt.go
│   │   └── wizard.go
│   ├── initializr/
│   │   ├── client.go
│   │   ├── metadata.go
│   │   └── request.go
│   ├── project/
│   │   ├── download.go
│   │   └── extract.go
│   └── validate/
│       └── validate.go
└── testdata/
    └── metadata.json
```

Rules: - Keep reusable implementation under `internal/`. -
`cmd/spring-create/main.go` should only start the app, report top-level
errors, and exit with the mapped exit code. - Do not create one giant
`main.go`. - Keep the module dependency-free. `go.mod` should contain no
`require` directives for third-party modules. - The exact module path
can differ if the owner chooses one, but package layout and behavior
should remain as specified.

## 5. Implementation sequence

Implement one phase at a time. After each phase, run formatting and
tests before moving forward.

### Phase 1 --- Project skeleton and CLI entry point

Files: `cmd/spring-create/main.go`, `internal/app/app.go`,
`internal/cli/args.go`

1.  Create the folders listed above.

2.  Implement argument parsing for `--help`, `--version`, `--url`,
    `--output`, `--no-color`, and `--verbose`.

3.  Reject unknown flags and missing flag values with usage text and
    exit code `2`.

4.  Put version in a variable set to `0.1.0`; structure the code so a
    release build can override it with `-ldflags`.

5.  Add top-level error-to-exit-code mapping.

6.  Add unit tests for valid flags, invalid flags, missing values, help,
    and version.

7.  Run:

    ``` powershell
    gofmt -w .
    go test ./...
    go vet ./...
    ```

Acceptance criteria: - `go run ./cmd/spring-create --help` prints usage
and exits `0`. - `go run ./cmd/spring-create --version` prints a version
and exits `0`. - Unknown flags return exit code `2`. - `go list -m all`
shows no third-party module dependencies.

### Phase 2 --- Input prompts and validation

Files: `internal/cli/prompt.go`, `internal/validate/validate.go`

1.  Implement reusable functions for:
    -   prompt with default,
    -   required string,
    -   numbered single-choice menu,
    -   comma-separated multi-choice menu,
    -   yes/no confirmation,
    -   cancellation (`q`/`quit`).
2.  Keep input/output streams injectable so tests can use
    `strings.Reader` and `bytes.Buffer`.
3.  Never call `log.Fatal` from reusable packages; return errors.
4.  Validate:
    -   group ID and package name using conservative Java identifier/dot
        rules;
    -   artifact ID and project name as non-empty safe names;
    -   description length to a documented limit;
    -   menu indexes and dependency IDs;
    -   output path is valid and resolvable.
5.  Preserve user input where valid. Do not silently transform invalid
    values.
6.  Add tests for blank/default input, whitespace, invalid menu indexes,
    multi-select duplicates, cancellation, invalid coordinates, and path
    validation.

Acceptance criteria: - Invalid choices cause a readable error and a
reprompt. - EOF is handled as cancellation/error, never an infinite
prompt loop. - Prompt code is testable without a real terminal.

### Phase 3 --- Spring Initializr metadata client

Files: `internal/initializr/client.go`,
`internal/initializr/metadata.go`, `testdata/metadata.json`

1.  Inspect the official metadata response:
    -   `GET https://start.spring.io/`
    -   Header: `Accept: application/vnd.initializr.v2.3+json`
    -   Header: `User-Agent: spring-create/0.1.0`
2.  Save a sanitized representative response to `testdata/metadata.json`
    for offline tests. Do not hardcode it as a production fallback.
3.  Define Go structs/custom decoding for the response's supported
    project types, languages, boot versions, Java versions, packaging,
    group/artifact defaults, and dependency groups.
4.  Inspect actual metadata field names and nesting before finalizing
    struct tags. Add fixtures for fields that may be absent or
    represented as arrays/objects.
5.  Implement an HTTP client with a 15-second metadata timeout.
6.  Require 2xx responses; cap metadata response body size (default 5
    MiB); close response bodies.
7.  Return clear errors for connection failures, non-2xx responses,
    invalid JSON, and missing required metadata.
8.  Add `httptest.Server` tests for success, HTTP error, malformed JSON,
    timeout, oversized response, and missing required metadata.
9.  Keep service base URL injectable so tests never call the real
    internet.

Acceptance criteria: - `go test ./...` performs no network requests. -
Production metadata is fetched live and used to build the menus. - No
hardcoded Spring Boot or Java version list is used as a fallback.

### Phase 4 --- Wizard and configuration model

Files: `internal/cli/wizard.go`, `internal/app/app.go`

1.  Define a `ProjectConfig` struct containing all project options and
    output path.
2.  Populate defaults from live metadata.
3.  Ask for project name first. Use it as the default artifact ID and
    base directory where valid; ask for group ID and package name with
    metadata-derived defaults.
4.  Use metadata-backed numbered menus for language, build type,
    packaging, Java version, and Boot version.
5.  Provide dependency selection by category:
    -   show category headings;
    -   show dependency display names and IDs;
    -   allow comma-separated selection of multiple entries;
    -   allow blank input for no optional dependencies;
    -   include a simple text filter/search prompt if feasible without
        adding dependencies.
6.  Keep the first version's dependency selector simple and usable. Do
    not implement fuzzy search, mouse input, or full-screen terminal
    rendering.
7.  Show a final summary of all selected values and dependencies.
8.  Require explicit `y/yes` before project generation. Anything else
    cancels.
9.  Test wizard flow with fixture metadata and fake input/output
    streams.

Acceptance criteria: - Every selection sent to the server is sourced
from metadata or validated user input. - User sees a summary and
confirms before download begins. - User can cancel from any prompt.

### Phase 5 --- Generation request and download

Files: `internal/initializr/request.go`, `internal/project/download.go`

1.  Build the request with `net/url.Values`, not string concatenation.
2.  Verify parameter names and multi-dependency encoding against the
    current official Spring Initializr documentation and an actual
    request.
3.  Request `/starter.zip` from the configured base URL.
4.  Apply a 60-second timeout to project generation.
5.  Limit response body to 100 MiB plus one byte for size detection;
    fail if limit is exceeded.
6.  Check HTTP status and content type where reliable. Do not assume
    content type alone proves that the response is a valid ZIP.
7.  Download to a temporary file, not directly into the final project
    directory.
8.  Check that the body is a valid ZIP before proceeding.
9.  Add tests using `httptest.Server` that returns a valid ZIP, non-2xx
    response, HTML error body, truncated ZIP, oversized response, and
    delayed response.
10. Do not follow redirects to non-HTTP(S) schemes. Use Go's standard
    redirect policy and validate any custom base URL.

Acceptance criteria: - Request encoding handles spaces, punctuation, and
multiple dependencies correctly. - Failed downloads leave no final
output directory behind. - Network timeouts and HTTP errors are readable
and actionable.

### Phase 6 --- Safe ZIP extraction

File: `internal/project/extract.go`

1.  Open the downloaded ZIP using `archive/zip`.
2.  Extract into a newly created temporary staging directory adjacent to
    the intended destination when possible, so final rename is on the
    same filesystem.
3.  For every entry:
    -   normalize separators;
    -   reject absolute paths, drive-letter paths, UNC paths, and
        traversal components;
    -   resolve the target and verify it remains within the staging
        root;
    -   reject symlink entries and other special file types;
    -   enforce a 250 MiB total uncompressed-size cap;
    -   enforce a reasonable maximum number of entries (default 20,000);
    -   create directories and files only inside the staging root;
    -   copy with a byte limit consistent with the declared and total
        limits;
    -   close each output file and ZIP entry reader, checking errors.
4.  Do not preserve dangerous permissions. Use conservative
    directory/file modes.
5.  Decide the final directory name from the validated project
    name/artifact ID and output parent directory. Explain the chosen
    final path before confirmation.
6.  If the final path exists, do not overwrite it. Ask the user to
    choose a different name/path; do not delete existing data.
7.  After all entries extract and validation succeeds, rename the
    staging directory to the final directory.
8.  On any failure, remove only the temporary staging directory created
    by this process.
9.  Add tests for valid archives, `../` traversal, absolute paths,
    Windows drive paths, symlinks, too many entries, oversized entries,
    existing target, and cleanup after error.

Acceptance criteria: - A malicious ZIP cannot write outside the staging
directory. - A failure never leaves a partially generated project at the
final path. - Existing user files are never silently overwritten.

### Phase 7 --- End-to-end behavior and usability

Files: `internal/app/app.go`, `README.md`

1.  Connect the wizard, metadata client, request builder, downloader,
    and extractor.
2.  Print concise progress messages:
    -   fetching metadata,
    -   downloading project,
    -   validating archive,
    -   extracting,
    -   done.
3.  Do not print progress spinners that require terminal escape codes.
4.  On success, print the absolute project directory.
5.  Print relevant next steps:
    -   Maven: `cd <project>` then `./mvnw spring-boot:run` on Unix-like
        systems, or `.\mvnw.cmd spring-boot:run` on Windows.
    -   Gradle: `cd <project>` then `.\gradlew.bat bootRun` on Windows
        or `./gradlew bootRun` on Unix-like systems.
    -   Mention that running/building the generated app requires a
        compatible JDK.
6.  Ensure output paths containing spaces work.
7.  Ensure `--url` works against an `httptest.Server` or local test
    service.
8.  Run the full test suite and a manual test against the real service.

Acceptance criteria: - One command starts the wizard. - A valid
configuration generates and extracts a project. - No Maven/Gradle/Java
command is executed automatically. - Errors explain what failed and how
the user can recover.

### Phase 8 --- Cross-platform build and release

1.  Build for Windows x86-64 first.
2.  Also verify Linux x86-64 and macOS ARM64 cross-compilation. The
    implementation must not use OS-specific terminal APIs in version 1.
3.  Add build instructions to README.
4.  Produce binaries with version metadata injected using `-ldflags`.
5.  Verify that builds do not require CGO.
6.  Keep dependencies at zero third-party Go modules.

Commands (run from project root in PowerShell):

``` powershell
# Format, test, and vet
gofmt -w .
go test ./...
go vet ./...

# Build for the current Windows machine
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags "-s -w -X main.version=0.1.0" -o .\dist\spring-create.exe .\cmd\spring-create

# Cross-compile Linux x86-64
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -trimpath -ldflags "-s -w -X main.version=0.1.0" -o .\dist\spring-create-linux-amd64 .\cmd\spring-create
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

# Verify dependencies
go list -m all
```

Important: create `dist` before building:

``` powershell
New-Item -ItemType Directory -Force .\dist
```

For macOS ARM64 cross-compilation:

``` powershell
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -trimpath -ldflags "-s -w -X main.version=0.1.0" -o .\dist\spring-create-darwin-arm64 .\cmd\spring-create
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

The version variable must be in package `main` in
`cmd/spring-create/main.go` for the exact `-X main.version=...` example
to work. If it is in a different package, update the linker flag
accordingly.

Acceptance criteria: - Windows executable builds and runs on a clean
Windows machine without Go installed. - Cross-compilation succeeds for
the documented targets. - `CGO_ENABLED=0` builds successfully. -
`go list -m all` lists only the main module and no third-party
dependencies.

## 6. Testing strategy

All automated tests must use only the standard library (`testing`,
`net/http/httptest`, `archive/zip`, `bytes`, `strings`, `os`,
`path/filepath`, and related standard packages).

Minimum test coverage:

-   CLI flags, usage, and version.
-   Input defaults, invalid selections, cancellation, EOF.
-   Project configuration validation.
-   Metadata parsing and missing fields.
-   HTTP timeout, bad status, invalid JSON, oversized metadata.
-   Query encoding and dependency list encoding.
-   ZIP validation and safe extraction.
-   Path traversal and Windows-specific path attacks.
-   Existing output directory behavior.
-   Cleanup after failed extraction/download.
-   End-to-end wizard using a fake Initializr HTTP server and in-memory
    input.

Do not use third-party test frameworks. Tests should be deterministic
and must not rely on live `start.spring.io`.

## 7. Definition of done

The project is done only when all of these are true:

-   [ ] `go test ./...` passes.
-   [ ] `go vet ./...` passes.
-   [ ] `gofmt -w .` has been run and produces no changes.
-   [ ] No third-party Go modules are imported or listed in `go.mod`.
-   [ ] CLI starts with no arguments and guides the user through the
    complete wizard.
-   [ ] Metadata and dependency choices come from the live Initializr
    service.
-   [ ] The generation request is encoded and verified against the
    official API.
-   [ ] Download size, timeout, HTTP status, ZIP path traversal, and
    uncompressed-size limits are enforced.
-   [ ] Existing directories are not overwritten.
-   [ ] Failed operations do not leave partial final projects.
-   [ ] Windows executable builds with `CGO_ENABLED=0`.
-   [ ] README explains installation, build, usage, troubleshooting, and
    security behavior.
-   [ ] Manual generation of a test project has been completed
    successfully.

## 8. README content required

The repository `README.md` must contain:

1.  What the tool does.
2.  Prerequisites for building (Go only).
3.  How to build from source on Windows PowerShell.
4.  How to run the generated executable.
5.  How to use `--help`, `--version`, `--url`, `--output`, and
    `--verbose`.
6.  Supported platforms and current limitations.
7.  Security model for downloading and extracting ZIP files.
8.  Troubleshooting for:
    -   `go` is not recognized (reopen terminal; check PATH; reinstall
        Go if necessary);
    -   no internet / proxy / TLS errors;
    -   Initializr metadata changes;
    -   target directory already exists;
    -   generated Spring project requires JDK and Maven/Gradle wrapper.
9.  State clearly that the CLI itself has no third-party Go
    dependencies.

## 9. Official references

Use these official references during implementation; verify the current
API before coding the metadata decoder and request builder.

-   Go download and installation: https://go.dev/dl/
-   Go installation instructions: https://go.dev/doc/install
-   Go build tutorial: https://go.dev/doc/tutorial/compile-install
-   Spring Initializr reference guide:
    https://docs.spring.io/initializr/docs/current/reference/html/
-   Spring Initializr source repository:
    https://github.com/spring-io/initializr
-   Production Initializr service: https://start.spring.io/

## 10. Instructions to the coding agent

Implement the project in the phases above, in order. Do not skip ahead
to a monolithic implementation.

At the end of each phase: 1. Run `gofmt -w .`. 2. Run `go test ./...`.
3. Run `go vet ./...`. 4. Fix failures before proceeding. 5. Report
which files changed, which tests ran, and any unresolved issues.

Before implementation, inspect the current Spring Initializr metadata
response and official API documentation. If the live response conflicts
with assumptions in this plan, adapt the decoder/request mapping to the
real response while preserving the user-facing requirements. Do not
invent metadata fields or request parameters.

Do not add third-party dependencies. If a requirement appears difficult
with the standard library, implement the smallest safe standard-library
solution or explain the blocker before changing scope. Do not replace
this plan with a different framework or language.
