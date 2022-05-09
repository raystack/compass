# Installation

## Requirements

Compass is written in Golang, and requires go version &gt;= 1.16. Please make sure that the go toolchain is available on your machine. See Golang’s [documentation](https://golang.org/) for installation instructions.

Alternatively, you can use docker to build Compass as a docker image. More on this in the next section.

Compass uses PostgreSQL 13 as its main storage and Elasticsearch v7 as the secondary storage to power the search. In order to run compass locally, you’ll need to have an instance of postgres and elasticsearch running. You can either download them and run it manually, or you can run them inside docker by using `docker-compose` with `docker-compose.yaml` provided in the root of this project and run the following command in a terminal

```text
$ docker-compose up
```
If you don't want to use `docker-compose`, you could run each storage (postgres and elasticsearch) individually with these commands

```text
$ docker run -d -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.6.1
```
```text
$ docker run -d -p 5432:5432 -e "POSTGRES_USER=compass" -e "POSTGRES_PASSWORD=compass_password" -e "POSTGRES_DB=compass" postgres:13
```

## Installation

Begin by cloning this repository then you have two ways in which you can build compass

* As a native executable
* As a docker image

To build compass as a native executable, run `make` inside the cloned repository.

```text
$ make
```

This will create the `compass` binary in the root directory

Building compass' Docker image is just a simple, just run docker build command and optionally name the image

```text
$ docker build . -t compass
```

