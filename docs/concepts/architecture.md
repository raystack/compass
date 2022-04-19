# Architecture

Compass' architecture is pretty simple. It serves HTTP server with Elasticsearch as its main persistent storage.

![Compass Architecture](../assets/architecture.jpg)

## System Design
### Components

#### HTTP Server

* HTTP server is the main and only interface to interact with Compass using RESTful pattern.

#### Elasticsearch

* Compass uses Elasticsearch as it is main storage for storing all of its metadata.
