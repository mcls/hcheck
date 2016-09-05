# hcheck

Poor man's health check. Intended to be put in your bashrc/zshrc/whatever.

## Install

```
go get https://github.com/mcls/hcheck
```

## Usage

```
usage: hcheck [flags] [urls]
  -errors-only
    only print errors
  -timeout duration
    request timeout in ms (default 500ms)
```

## Example

```text
$ hcheck -timeout 1000ms \
    https://github.com \
    https://google.com \
    https://www.facebook.com
200 OK (293ms) - https://www.facebook.com
200 OK (333ms) - https://google.com
200 OK (423ms) - https://github.com
```
