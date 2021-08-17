# `gyv` - The official Goyave CLI

[![Version](https://img.shields.io/github/v/release/go-goyave/gyv?include_prereleases)](https://github.com/go-goyave/gyv/releases)
[![Build Status](https://github.com/go-goyave/gyv/workflows/Test/badge.svg)](https://github.com/go-goyave/gyv/actions)
[![Coverage Status](https://coveralls.io/repos/github/go-goyave/gyv/badge.svg)](https://coveralls.io/github/go-goyave/gyv)

## 🚧 Work in progress

The official CLI for the [Goyave](https://github.com/go-goyave/goyave) REST API framework.

- Project creation
- Scaffolding and quick prototyping
- Utility: seeders, migrations, routes list and more

## Install

**Minimum Go version:** 1.16

```
go install goyave.dev/gyv@latest
```

## Usage

```sh
# Create a new project
gyv create project

# Create a new controller named "hello"
gyv create controller --name "hello"

# Create a new model named "User"
gyv create model --name "user"

# Create a new middleware named "Auth"
gyv create middleware --name "auth"

# Generate OpenAPI3 specification of your application
gyv openapi
```

## License

`gyv` is MIT Licensed. Copyright (c) 2021 Jérémy LAMBERT (SystemGlitch) and Louis LAURENT (ulphidius)