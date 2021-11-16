# Columbus

![test workflow](https://github.com/odpf/columbus/actions/workflows/test.yml/badge.svg)
![build workflow](https://github.com/odpf/columbus/actions/workflows/build.yml/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/columbus?logo=semantic-release)](Version)

Columbus is a search and discovery engine built for querying application deployments, datasets and meta resources. It can also optionally track data flow relationships between these resources and allow the user to view a representation of the data flow graph.

<p align="center"><img src="./docs/assets/overview.svg" /></p>

## Key Features
Discover why users choose Columbus as their main data discovery and lineage service

* **Full text search** Faster and better search results powered by ElasticSearch full text search capability.
* **Search Tuning** Narrow down your search results by adding filters, getting your crisp results.
* **Data Lineage** Understand the relationship between metadata with data lineage interface.
* **Scale:** Columbus scales in an instant, both vertically and horizontally for high performance.
* **Extensibility:** Add your own metadata types and resources to support wide variety of metadata.
* **Runtime:** Columbus can run inside VMs or containers in a fully managed runtime environment like kubernetes.

## Usage

Explore the following resources to get started with Columbus:

* [Guides](docs/guides) provides guidance on ingesting and queying metadata from Columbus.
* [Concepts](docs/concepts) describes all important Columbus concepts.
* [Reference](docs/reference) contains details about configurations, metrics and other aspects of Columbus.
* [Contribute](docs/contribute/contribution.md) contains resources for anyone who wants to contribute to Columbus.

## Requirements

Columbus is written in golang, and requires go version >= 1.16. Please make sure that the go tool chain is available on your machine. See golang’s [documentation](https://golang.org/) for installation instructions.

Alternatively, you can use docker to build columbus as a docker image. More on this in the next section.

Columbus uses elasticsearch v7 as the query and storage backend. In order to run columbus locally, you’ll need to have an instance of elasticsearch running.  You can either download elasticsearch and run it manually, or you can run elasticsearch inside docker by running the following command in a terminal
```
$ docker run -d -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.6.1
```

## Running locally
Begin by cloning this repository, then you have two ways in which you can build columbus
* As a native executable
* As a docker image

To build columbus as a native executable, run `make` inside the cloned repository.
```
$ make
```

This will create the `columbus` binary in the root directory

Building columbus’s Docker image is just a simple, just run docker build command and optionally name the image
```
$ docker build . -t columbus
```

Columbus interfaces with an elasticsearch cluster. Run columbus using:

```
./columbus -elasticsearch-brokers "http://<broker-host-name>"
```

Elasticsearch brokers can alternatively be specified via the `ELASTICSEARCH_BROKERS` environment variable.

If you used Docker to build columbus, then configuring networking requires extra steps. Following is one of doing it, running elasticsearch inside docker

```
# create a docker network where columbus and elasticsearch will reside 
$ docker network create columbus-net

# run elasticsearch, bound to the network we created. Since we are using the -d flag to docker run, the command inside the subshell returns the container id
$ ES_CONTAINER_ID=$(docker run -d -e "discovery.type=single-node" --net columbus-net elasticsearch:7.5.2)

# run columbus, passing in the hostname (container id) of the elasticsearch server
# if everything goes ok, you should say something like this:

# time="2020-04-01T18:41:00Z" level=info msg="columbus v0.1.0-103-g83b909b starting on 0.0.0.0:8080" reporter=main
# time="2020-04-01T18:41:00Z" level=info msg="connected to elasticsearch cluster \"docker-cluster\" (server version 7.5.2)" reporter=main
$ docker run --net columbus-net columbus -p 8080:8080 -elasticsearch-brokers http://${ES_CONTAINER_ID}:9200 
```

## Running tests

```
# Run unit tests
$ make unit-test

# Run integration tests
$ make test
```

The integration test suite requires docker to run elasticsearch. In case you wish to test against an existing 
elasticsearch cluster, set the value of `ES_TEST_SERVER_URL` to the URL of the elasticsearch server.


## Contribute

Development of Columbus happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and improvements. Read below to learn how you can take part in improving Columbus.

Read our [contributing guide](docs/contribute/contribution.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Columbus.

To help you get your feet wet and get you familiar with our contribution process, we have a list of [good first issues](https://github.com/odpf/columbus/labels/good%20first%20issue) that contain bugs which have a relatively limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/columbus/graphs/contributors).

## License
Columbus is [Apache 2.0](LICENSE) licensed.