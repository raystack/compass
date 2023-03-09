# Compass

![test workflow](https://github.com/odpf/compass/actions/workflows/test.yml/badge.svg)
![build workflow](https://github.com/odpf/compass/actions/workflows/build_dev.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/odpf/compass/badge.svg?branch=main)](https://coveralls.io/github/odpf/compass?branch=main)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/compass?logo=semantic-release)](Version)

Compass is a search and discovery engine built for querying application deployments, datasets and meta resources. It can also optionally track data flow relationships between these resources and allow the user to view a representation of the data flow graph.

<p align="center"><img src="./docs/static/assets/overview.svg" /></p>

## Key Features

Discover why users choose Compass as their main data discovery and lineage service

- **Full text search** Faster and better search results powered by ElasticSearch full text search capability.
- **Search Tuning** Narrow down your search results by adding filters, getting your crisp results.
- **Data Lineage** Understand the relationship between metadata with data lineage interface.
- **Scale:** Compass scales in an instant, both vertically and horizontally for high performance.
- **Extensibility:** Add your own metadata types and resources to support wide variety of metadata.
- **Runtime:** Compass can run inside VMs or containers in a fully managed runtime environment like kubernetes.

## Documentation

Explore the following resources to get started with Compass:

- [Guides](https://odpf.github.io/compass/docs/guides) provides guidance on ingesting and querying metadata from Compass.
- [Concepts](https://odpf.github.io/compass/docs/concepts) describes all important Compass concepts.
- [Reference](https://odpf.github.io/compass/docs/reference) contains details about configurations, metrics and other aspects of Compass.
- [Contribute](https://odpf.github.io/compass/docs/contribute/contribution.md) contains resources for anyone who wants to contribute to Compass.

## Installation

Install Compass on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. <br/>Refer this for [installations](https://odpf.github.io/compass/docs/installation) and [configurations](https://odpf.github.io/compass/docs/guides/configuration)

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/compass/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`compass` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/odpf/compass/releases/latest) page:

```sh
brew install odpf/tap/compass
```

To upgrade to the latest version:

```
brew upgrade compass
```

Check for installed compass version

```sh
compass version
```

#### Linux

`compass` is available as downloadable binaries from the [releases](https://github.com/odpf/compass/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`compass` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/odpf/compass/releases/latest) page:

```
scoop bucket add compass https://github.com/odpf/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update compass
```

#### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull odpf/compass:latest
```

To pull a specific version:

```
docker pull odpf/compass:v0.3.2
```

If you like to have a shell alias that runs the latest version of compass from docker whenever you type `compass`:

```
mkdir -p $HOME/.config/odpf
alias compass="docker run -e HOME=/tmp -v $HOME/.config/odpf:/tmp/.config/odpf --user $(id -u):$(id -g) --rm -it -p 3306:3306/tcp odpf/compass:latest"
```

## Usage

Compass is purely API-driven. It is very easy to get started with Compass. It provides CLI, HTTP and GRPC APIs for simpler developer experience.

#### CLI

Compass CLI is fully featured and simple to use, even for those who have very limited experience working from the command line. Run `compass --help` to see list of all available commands and instructions to use.

List of commands

```
compass --help
```

Print command reference

```sh
compass reference
```

#### API

Compass provides a fully-featured GRPC and HTTP API to interact with Compass server. Both APIs adheres to a set of standards that are rigidly followed. Please refer to [proton](https://github.com/odpf/proton/tree/main/odpf/compass/v1beta1) for GRPC API definitions.

<details>
  <summary>Dependencies:</summary>

- Compass is written in Golang, and requires go version &gt;= 1.16. Please make sure that the go toolchain is available on your machine. See Golang’s [documentation](https://golang.org/) for installation instructions. Alternatively, you can use docker to build Compass as a docker image. More on this in the next section.
- Compass uses PostgreSQL 13 as its main storage and Elasticsearch v7 as the secondary storage to power the search. In order to run compass locally, you’ll need to have an instance of postgres and elasticsearch running. You can either download them and run it manually, or you can run them inside docker by using `docker-compose` with `docker-compose.yaml` provided in the root of this project.
- PostgreSQL details and Elasticsearch brokers can alternatively be specified via the environment variable, `ELASTICSEARCH_BROKERS` for elasticsearch and `DB_HOST`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` for postgres.
- If you use Docker to build compass, then configuring networking requires extra steps. Following is one of doing it by running postgres and elasticsearch inside with `docker-compose` first.

- Go to the root of this project and run `docker-compose`.

  ```text
  docker-compose up
  ```

- Once postgres and elasticsearch has been ready, we can run Compass by passing in the config of postgres and elasticsearch defined in `docker-compose.yaml` file.
</details>

## Building Compass

Begin by cloning this repository then you have two ways in which you can build compass

- As a native executable
- As a docker image

To build compass as a native executable, run `make` inside the cloned repository.

```text
make
```

This will create the `compass` binary in the root directory

Building compass' Docker image is just a simple, just run docker build command and optionally name the image

```text
docker build . -t compass
```

## Migration

Before serving Compass app, we need to run the migration first. Run this docker command to migrate Compass.

```text
$ docker run --rm --net compass_storage -p 8080:8080 -e ELASTICSEARCH_BROKERS=http://es:9200 -e DB_HOST=postgres -e DB_PORT=5432 -e DB_NAME=compass -e DB_USER=compass -e DB_PASSWORD=compass_password odpf/compass compass server migrate
```

If you are using Compass binary, you can run this command.

```text
./compass -elasticsearch-brokers "http://<broker-host-name>" -db-host "<postgres-host-name>" -db-port 5432 -db-name "<postgres-db-name>" -db-user "<postgres-db-user>" -db-password "<postgres-db-password> server migrate"
```

## Serving locally

Once the migration has been done, Compass server can be started with this command.

```text
docker run --net compass_storage -p 8080:8080 -e ELASTICSEARCH_BROKERS=http://es:9200 -e DB_HOST=postgres -e DB_PORT=5432 -e DB_NAME=compass -e DB_USER=compass -e DB_PASSWORD=compass_password odpf/compass compass server start
```

If you are using Compass binary, you can run this command.

```text
./compass -elasticsearch-brokers "http://<broker-host-name>" -db-host "<postgres-host-name>" -db-port 5432 -db-name "<postgres-db-name>" -db-user "<postgres-db-user>" -db-password "<postgres-db-password> server start"
```

## Running tests

Running all unit tests

```
make test
```

The tests combine both unit and integration tests, the test suite requires docker to run elasticsearch. In case you wish to test against an existing
elasticsearch cluster, set the value of `ES_TEST_SERVER_URL` to the URL of the elasticsearch server.

## Contribute

Development of Compass happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and improvements. Read below to learn how you can take part in improving Compass.

Read our [contributing guide](https://odpf.github.io/compass/docs/contribute/contribution.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Compass.

To help you get your feet wet and get you familiar with our contribution process, we have a list of [good first issues](https://github.com/odpf/compass/labels/good%20first%20issue) that contain bugs which have a relatively limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/compass/graphs/contributors).

## License

Compass is [Apache 2.0](LICENSE) licensed.
