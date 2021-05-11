# confx

**`confx`** is a [go(golang)](http://golang.org/) package for application configuration management.

## Features

* Provide a mechanism to set override values (highest precedence).
* Load and unmarshal a configuration file in YAML.
* Provide a mechanism to set default values (lowest precedence).

## Installation

> **Note:** `confx` uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies.

```bash
▶ go get github.com/signedsecurity/confx
```
## Usage

### Set Configuration

### Set Overrides

### Set Defaults

### Getting Values

In `confx`, there are a few ways to get a value depending on the value’s type:

| Functions               | Return Type   |
| :---------------------- | :------------ |
| `Get(key string)`       | `interface{}` |
| `GetInt(key string)`    | `int`         |
| `GetString(key string)` | `string`      |