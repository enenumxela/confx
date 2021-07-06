# confx

**`confx`** is a [go(golang)](http://golang.org/) package for application configuration management.

## Resources

* [Featuers](#features)
* [Installation](#installation)
* [Usage](#usage)
    * [Setting Values](#setting-values)
        * [Set Configuration](#set-configuration)
        * [Set Overrides](#set-overrides)
        * [Set Defaults](#set-defaults)
    * [Getting Values](#getting-values)
        * [Get Nested Values](#get-nested-values)

## Features

* Provide a mechanism to set override values (highest precedence).
* Load and unmarshal a YAML configuration file.
* Provide a mechanism to set default values (lowest precedence).

## Installation

**Note:** `confx` uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage its dependencies.

```bash
go get github.com/enenumxela/confx
```
## Usage

### Setting Values

#### Set Configuration

Examples:

```go
if err := confx.SetConfiguration("./conf.yaml"); err != nil {
    log.Fatalln(err)
}
```

#### Set Overrides

Examples:

```go
confx.SetOverride("override", "override value")
```

#### Set Defaults

Default values are not required, but are useful in the event that a key hasn't been set via config file.

Examples:

```go
confx.SetDefault("default", "default value")
```

### Getting Values

In `confx`, there are a few ways to get a value depending on the value’s type:

| Functions               | Return Type   |
| :---------------------- | :------------ |
| `Get(key string)`       | `interface{}` |
| `GetInt(key string)`    | `int`         |
| `GetString(key string)` | `string`      |

#### Get Nested Values

confx can access a nested value by passing a `.` delimited path of keys:

```go
Get("datastore.metric.host")
```