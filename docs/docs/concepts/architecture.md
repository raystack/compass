# Architecture

Compass' architecture is pretty simple. It has a client-server architecture backed by PostgreSQL as a main storage and Elasticsearch as a secondary storage and provides HTTP & gRPC interface to interact with.

![Compass Architecture](/assets/architecture.png)

## System Design

### Components

#### gRPC Server

- gRPC server is the main interface to interact with Compass.
- The protobuf file to define the interface is centralized in [raystack/proton](https://github.com/raystack/proton/tree/main/raystack/compass/v1beta1)

#### gRPC-gateway Server

- gRPC-gateway server transcodes HTTP call to gRPC call and allows client to interact with Compass using RESTful HTTP request.

#### PostgreSQL

- Compass uses PostgreSQL as it is main storage for storing all of its metadata.

#### Elasticsearch

- Compass uses Elasticsearch as it is secondary storage to power search of metadata.
