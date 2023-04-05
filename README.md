# CI Analysis Tool [![Go Report Card](https://goreportcard.com/badge/github.com/ashishb/citool)](https://goreportcard.com/report/github.com/ashishb/citool) [![Test](https://github.com/ashishb/citool/actions/workflows/test.yml/badge.svg)](https://github.com/ashishb/citool/actions/workflows/test.yml)

[![Lint YAML](https://github.com/ashishb/citool/actions/workflows/lint-yaml.yaml/badge.svg)](https://github.com/ashishb/citool/actions/workflows/lint-yaml.yaml)
[![Lint Go](https://github.com/ashishb/citool/actions/workflows/lint-go.yaml/badge.svg)](https://github.com/ashishb/citool/actions/workflows/lint-go.yaml)
[![Validate Go code formatting](https://github.com/ashishb/citool/actions/workflows/format-go.yaml/badge.svg)](https://github.com/ashishb/citool/actions/workflows/format-go.yaml)

A tool to download and analyze Circle CI jobs data.

## Installations

`go get -u github.com/ashishb/citool`

## CLI arguments

```
Usage of ./citool:
  -branch string
    Optional branch name to filter download/analysis on
  -circle-token string
    Circle CI access token. Download mode only.
  -debug
    Set this to true to enable debug logging
  -download-dir string
    Directory to download Circle CI data to (default "./circleci_data")
  -input-files string
    Comma-separated list of files containing downloaded job results from CircleCI. Analyze mode only.
  -jobname string
    Only consider job results for this jobname. Analyze mode only.
  -jobstatus string
    Only consider job results with this completion status. Analyze mode only.
  -limit int
    Circle CI build results download limit
  -mode string
    Mode - "download" or "analyze"
  -offset int
    Circle CI build results download start offset
  -print-duration
    Print per-job average duration. Analyze mode only. (default true)
  -print-duration-graph
    Print per-job duration time series graph (yes, a graph). Analyze mode only. (default true)
  -print-success-graph
    Print per-job success graph (yes, a graph). Analyze mode only. (default true)
  -print-success-rate
    Print per-job aggregated success rate. Analyze mode only. (default true)
  -reponame string
    Optional repository name to filter downloads/analysis on
  -username string
    Optional username to filter downloads/analysis on
  -vcsType string
    Name of the VCS system - See https://circleci.com/docs/api/#version-control-systems-vcs-type. Download mode only. (default "github")
  -version
    Prints version of this tool
```

## Examples

Generate a Circle CI API token at [https://circleci.com/account/api](https://circleci.com/account/api). Use the token to **download the data**

```
./citool --mode download --circle-token ${TOKEN} --limit 100 --offset 0 --username ashishb --reponame androidtool --download-dir androidtool_data
```

Now, analyze

```
$ ./citool --mode analyze androidtool_data/*.json
Number of job results: 100
Job name    Success Rate
--------    -----------
Build Error 0/1 (0%)
build-jdk8  34/43 (79%)
build-jdk9  39/47 (82%)

Job name    Average job duration
----------  --------------------
build-jdk8  37s
build-jdk9  23s
Build Error 1s


Job name: build-jdk8 (25 data points)

 90.48 ┤                                                                     ╭╮
 89.96 ┤                                                                     │╰╮
 89.44 ┤                                                                    ╭╯ ╰─╮
 88.92 ┤                                                                    │    ╰──╮
 88.40 ┤                                                                   ╭╯       ╰─────╮
 87.88 ┤                                             ╭╮                    │              ╰╮
 87.36 ┤                                            ╭╯╰──╮      ╭─╮        │               ╰╮
 86.84 ┤                                           ╭╯    ╰──╮ ╭─╯ ╰─╮     ╭╯                │
 86.31 ┤                    ╭─╮                   ╭╯        ╰─╯     ╰╮    │                 ╰╮
 85.79 ┤                   ╭╯ ╰─╮                 │                  ╰────╯                  │       ╭──╮
 85.27 ┤                  ╭╯    ╰╮               ╭╯                                          ╰╮    ╭─╯  ╰╮
 84.75 ┤                 ╭╯      │             ╭─╯                                            ╰────╯     ╰─
 84.23 ┤                ╭╯       │            ╭╯
 83.71 ┼╮         ╭─────╯        ╰╮       ╭───╯
 83.19 ┤╰╮     ╭──╯               │     ╭─╯
 82.67 ┤ ╰─╮ ╭─╯                  ╰╮    │
 82.15 ┤   ╰─╯                     │   ╭╯
 81.62 ┤                           │  ╭╯
 81.10 ┤                           ╰╮ │
 80.58 ┤                            │╭╯
 80.06 ┤                            ╰╯

Job name: build-jdk9 (30 data points)

 56.36 ┼──╮
 55.47 ┤  ╰─╮       ╭╮
 54.58 ┤    ╰───────╯╰╮
 53.69 ┤              ╰╮
 52.80 ┤               ╰╮
 51.91 ┤                │
 51.02 ┤                ╰────╮                                      ╭─╮           ╭──╮
 50.13 ┤                     ╰╮                                    ╭╯ ╰─╮         │  ╰───╮                ╭
 49.24 ┤                      ╰╮                                  ╭╯    ╰╮        │      ╰──╮             │
 48.35 ┤                       ╰╮                                 │      ╰╮       │         ╰─╮          ╭╯
 47.46 ┤                        │                                ╭╯       │      ╭╯           ╰─╮       ╭╯
 46.57 ┤                        │                              ╭─╯        ╰╮     │              ╰╮      │
 45.68 ┤                        ╰╮                        ╭─╮ ╭╯           ╰╮    │               ╰─╮  ╭─╯
 44.79 ┤                         │                      ╭─╯ ╰─╯             ╰╮  ╭╯                 ╰──╯
 43.90 ┤                         ╰╮                   ╭─╯                    │  │
 43.01 ┤                          │                  ╭╯                      ╰╮ │
 42.12 ┤                          │                 ╭╯                        ╰─╯
 41.23 ┤                          ╰╮               ╭╯
 40.33 ┤                           ╰╮             ╭╯
 39.44 ┤                            ╰─╮   ╭───────╯
 38.55 ┤                              ╰───╯

```

```
$ ./citool --version
0.1.0
```

### Development

1. `make citool` - to build
1. `make test` - to test
1. `make lint` - to vet and lint
1. `make format` - to format using gofmt
