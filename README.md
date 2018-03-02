# AUXO

**auxo** is an all-in-one Go framework for simplifying program development.

> **WARNING**: This package is a work in progress. It's API may still break in backwards-incompatible ways without warnings. Please use dependency management tools such as **[dep](https://github.com/golang/dep)** to lock version.

## Components

* **CLI** - Easily build a friendly CLI program with sub commands support
* **Config** - Manage configuration for various formats and sources
* **Log** - Flexible **log4j** style log component
* **Cache** - Simple and elegant cache management
* **Web** - Web server with a variety of advanced features
* **RPC** - Lightweight and high performace
    * Service Discovery - Automatic registration and name resolution with service discovery
    * Load Balancing - Smart client side load balancing of services built on discovery
* **Database**
    * GSD - A lightweight, fluent SQL data access and ORM library
    * MongoDB - A powerful wrapper for [mgo](https://github.com/globalsign/mgo)
* **Utility** - Some useful utility packages...

## Goals

* Simple and elegant API
* Easy to use and maintain
* Focus on performance

## Installation

* Using `go get`

```bash
> go get -u github.com/cuigh/auxo
```

* Using `dep`

```bash
> cd <PATH/TO/PROJECT>
> dep ensure -add github.com/cuigh/auxo
```

## Documentation

* **[English](https://cuigh.tech/auxo/)**
* **[中文](https://cuigh.tech/auxo/zh/)**(TODO)
