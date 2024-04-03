[![Build Status](https://github.com/TestRecall/reporter/workflows/CI/badge.svg?branch=main)](https://github.com/TestRecall/reporter/actions?query=workflow%3ACI)
[![GitHub release](https://img.shields.io/github/release/TestRecall/reporter.svg)](https://github.com/TestRecall/reporter/releases/latest)
[![license](https://img.shields.io/github/license/TestRecall/reporter.svg)](https://github.com/TestRecall/reporter/blob/master/LICENSE)

# TestRecall Reporter

This is a TestRecall command line for uploading test reports.

- Downloads can be found under [releases][releases_url]
- Documentation can be found [here][docs_url]

## Installation

There are a few options for installing

### Recommended way:

```bash
curl -fsSL https://get.testrecall.com/reporter | bash

testrecall-reporter --version
```

### From the latest [releases][releases_url]:

```bash
curl -fsSL -o testrecall-reporter \
  https://github.com/TestRecall/reporter/releases/download/v0.2.0/reporter_Linux_x86_64

chmod +x testrecall-reporter

./testrecall-reporter --version
```

### From source:

```bash
go get github.com/golangci/golangci-lint/cmd/golangci-lint@v0.2.0

reporter --version
```

## Usage

The TestRecall reporter uploads test results from your test suites. If your
language can output a junit xml format test reports, running `testrecall-reporter`
after your test results will upload the results.

```bash
TR_UPLOAD_TOKEN=your_upload_token

trap 'testrecall-reporter' EXIT

npm run test # => output report.xml
```

### Configuration

| flag    | environment       | values                             | note                                                                                         |
| ------- | ----------------- | ---------------------------------- | -------------------------------------------------------------------------------------------- |
| `file`  |                   | \<glob\>                           | file path or glob pattern for xml results, e.g. (`/tmp/report.xml`, or `build/*/junit*.xml`) |
| `multi` |                   | \<'before' \| 'multi' \| 'after'\> | enables multi-stage uploads, for suites that execute on multiple runners                     |
|         | `TR_UPLOAD_TOKEN` | \<string\>                         | upload token for your test project                                                           |

The test reporter will pick up most configuration options by default, including common default locations for test reports.

## Compiling

If you want to compile from source, you will need:

- Go
- Make

To build, test, and compile a binary:

```bash
make setup
make test
make build
```

[releases_url]: https://github.com/TestRecall/reporter/releases
[docs_url]: https://testrecall.com/docs
