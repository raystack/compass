# Development Guide

This guide is intended for developers who want to contribute to Compass. It contains information on how to build, test and run Compass locally.

## Building Compass

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
$ docker run --rm --net compass_storage -p 8080:8080 -e ELASTICSEARCH_BROKERS=http://es:9200 -e DB_HOST=postgres -e DB_PORT=5432 -e DB_NAME=compass -e DB_USER=compass -e DB_PASSWORD=compass_password raystack/compass compass server migrate
```

If you are using Compass binary, you can run this command.

```text
./compass -elasticsearch-brokers "http://<broker-host-name>" -db-host "<postgres-host-name>" -db-port 5432 -db-name "<postgres-db-name>" -db-user "<postgres-db-user>" -db-password "<postgres-db-password> server migrate"
```

## Serving locally

Once the migration has been done, Compass server can be started with this command.

```text
docker run --net compass_storage -p 8080:8080 -e ELASTICSEARCH_BROKERS=http://es:9200 -e DB_HOST=postgres -e DB_PORT=5432 -e DB_NAME=compass -e DB_USER=compass -e DB_PASSWORD=compass_password raystack/compass compass server start
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
