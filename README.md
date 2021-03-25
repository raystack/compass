# Columbus

Columbus is a search server for querying application deployments, datasets and schemas.

## Requirements
Columbus is written in golang, and requires go version >= 1.13. Please make sure that the go tool chain is available on your machine. See golang’s [documentation](https://golang.org/) for installation instructions.

Alternatively, you can use docker to build columbus as a docker image. More on this in the next section.

Columbus uses elasticsearch v7 as the query and storage backend. In order to run columbus locally, you’ll need to have an instance of elasticsearch running.  You can either download elasticsearch and run it manually, or you can run elasticsearch inside docker by running the following command in a terminal
```
$ docker run -d -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.6.1
```

## Installation
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

## Usage
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

## Testing

Unit tests can be run using:
```
make unit-test
```

and integration (+ unit) tests with
```
make test
```

The integration test suite requires docker to run elasticsearch. In case you wish to test against an existing 
elasticsearch cluster, set the value of `ES_TEST_SERVER_URL` to the URL of the elasticsearch server.
