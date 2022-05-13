# Installation

## Requirements

Compass is written in Golang, and requires go version &gt;= 1.16. Please make sure that the go toolchain is available on your machine. See Golang’s [documentation](https://golang.org/) for installation instructions.

Alternatively, you can use docker to build Compass as a docker image. More on this in the next section.

Compass uses PostgreSQL 13 as its main storage and Elasticsearch v7 as the secondary storage to power the search. In order to run compass locally, you’ll need to have an instance of postgres and elasticsearch running. You can either download them and run it manually, or you can run them inside docker by using `docker-compose` with `docker-compose.yaml` provided in the root of this project.

PostgreSQL details and Elasticsearch brokers can alternatively be specified via the environment variable, `ELASTICSEARCH_BROKERS` for elasticsearch and `DB_HOST`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` for postgres.

If you use Docker to build compass, then configuring networking requires extra steps. Following is one of doing it by running postgres and elasticsearch inside with `docker-compose` first.

Go to the root of this project and run `docker-compose`.

```bash
$ docker-compose up
```
Once postgres and elasticsearch has been ready, we can run Compass by passing in the config of postgres and elasticsearch defined in `docker-compose.yaml` file.

## Building Compass

Begin by cloning this repository then you have two ways in which you can build compass

* As a native executable
* As a docker image

To build compass as a native executable, run `make` inside the cloned repository.

```bash
$ make
```

This will create the `compass` binary in the root directory

Building compass' Docker image is just a simple, just run docker build command and optionally name the image

```bash
$ docker build . -t compass
```

## Migration
Before serving Compass app, we need to run the migration first. Run this docker command to migrate Compass.

```bash
$ docker run --rm --net compass_storage -p 8080:8080 -e ELASTICSEARCH_BROKERS=http://es:9200 -e DB_HOST=postgres -e DB_PORT=5432 -e DB_NAME=compass -e DB_USER=compass -e DB_PASSWORD=compass_password odpf/compass compass migrate
```

If you are using Compass binary, you can run this command.
```bash
./compass -elasticsearch-brokers "http://<broker-host-name>" -db-host "<postgres-host-name>" -db-port 5432 -db-name "<postgres-db-name>" -db-user "<postgres-db-user>" -db-password "<postgres-db-password> migrate"
```

## Serving

Once the migration has been done, Compass server can be started with this command.

```bash
$ docker run --net compass_storage -p 8080:8080 -e ELASTICSEARCH_BROKERS=http://es:9200 -e DB_HOST=postgres -e DB_PORT=5432 -e DB_NAME=compass -e DB_USER=compass -e DB_PASSWORD=compass_password odpf/compass compass serve
```

If you are using Compass binary, you can run this command.
```bash
./compass -elasticsearch-brokers "http://<broker-host-name>" -db-host "<postgres-host-name>" -db-port 5432 -db-name "<postgres-db-name>" -db-user "<postgres-db-user>" -db-password "<postgres-db-password> serve"
```

If everything goes ok, you should see something like this:
```bash
time="2022-04-27T09:18:08Z" level=info msg="compass starting" version=v0.2.0
time="2022-04-27T09:18:08Z" level=info msg="connected to elasticsearch cluster" config="\"docker-cluster\" (server version 7.6.1)"
time="2022-04-27T09:18:08Z" level=info msg="New Relic monitoring is disabled."
time="2022-04-27T09:18:08Z" level=info msg="statsd metrics monitoring is disabled."
time="2022-04-27T09:18:08Z" level=info msg="connected to postgres server" host=postgres port=5432
time="2022-04-27T09:18:08Z" level=info msg="server started"
```

## Required Header/Metadata in API
Compass has a concept of [User](../concepts/user.md). In the current version, all HTTP & gRPC APIs in Compass requires an identity header/metadata in the request. The header key is configurable but the default name is `Compass-User-UUID`.

Compass APIs also expect an additional optional e-mail header. This is also configurable and the default name is `Compass-User-Email`. The purpose of having this optional e-mail header is described in the [User](../concepts/user.md) section.