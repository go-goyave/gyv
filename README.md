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

```
gyv create project
gyv create controller --controller-name "hello"
gyv create model --model-name "user"
gyv create middleware --middleware-name "auth"
```

## License

`gyv` is MIT Licensed. Copyright (c) 2021 Jérémy LAMBERT (SystemGlitch) and Louis LAURENT (ulphidius)