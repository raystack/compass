# Installation

## Requirements

Compass is written in Golang, and requires go version &gt;= 1.16. Please make sure that the go toolchain is available on your machine. See Golang’s [documentation](https://golang.org/) for installation instructions.

Alternatively, you can use docker to build Compass as a docker image. More on this in the next section.

Compass uses elasticsearch v7 as the query and storage backend. In order to run compass locally, you’ll need to have an instance of elasticsearch running. You can either download elasticsearch and run it manually, or you can run elasticsearch inside docker by running the following command in a terminal

```text
$ docker run -d -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.6.1
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

