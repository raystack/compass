# Compass

![test workflow](https://github.com/raystack/compass/actions/workflows/test.yml/badge.svg)
![release workflow](https://github.com/raystack/compass/actions/workflows/release.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/raystack/compass/badge.svg?branch=main)](https://coveralls.io/github/raystack/compass?branch=main)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/raystack/compass?logo=semantic-release)](Version)

Compass is a context engine that builds a knowledge graph of your organization's metadata, capturing entities, relationships, and lineage across systems and time, making it discoverable and queryable for both humans and AI agents.

Critical organizational knowledge lives scattered across dozens of systems: services, datasets, applications, teams, configurations, decisions, and the relationships between them. Compass resolves observations from these sources into unified entities, constructs a temporal graph of their relationships, and indexes everything for both keyword and semantic search. The result is a context graph that stitches together what exists, who owns it, how it connects, and what changed over time, so both humans and AI agents can discover, traverse, and reason over the full picture.

<p align="center"><img src="./docs/static/assets/overview.svg" /></p>

## Key Features

- **Entity Resolution:** Resolve and deduplicate metadata observations from multiple sources into unified entities with stable identity.
- **Knowledge Graph:** Store typed, directed relationships between entities including lineage, ownership, documentation, and custom edge types.
- **Hybrid Search:** Combine keyword precision with semantic similarity using Postgres-native full-text search and pgvector embeddings.
- **Graph Traversal:** Multi-hop traversal queries across the entity graph for impact analysis, dependency tracking, and path discovery.
- **Context Composition:** Assemble schema, lineage, ownership, and quality signals into context documents ready for LLM consumption.
- **AI Serving:** Expose the full graph as an MCP server so AI agents can discover, traverse, and reason over organizational knowledge.
- **Extensibility:** Open type system for entities and relationships to support any kind of metadata across your infrastructure.

## Documentation

Explore the following resources to get started with Compass:

- [Guides](https://compass-raystack.vercel.app/docs/guides) provides guidance on ingesting and querying metadata from Compass.
- [Concepts](https://compass-raystack.vercel.app/docs/concepts) describes all important Compass concepts.
- [Reference](https://compass-raystack.vercel.app/docs/reference) contains details about configurations, metrics and other aspects of Compass.
- [Contribute](https://compass-raystack.vercel.app/docs/contribute/contribution.md) contains resources for anyone who wants to contribute to Compass.

## Installation

Install Compass on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. <br/>Refer this for [installations](https://compass-raystack.vercel.app/docs/installation) and [configurations](https://compass-raystack.vercel.app/docs/guides/configuration)

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/raystack/compass/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`compass` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/raystack/compass/releases/latest) page:

```sh
brew install raystack/tap/compass
```

To upgrade to the latest version:

```
brew upgrade compass
```

#### Linux

`compass` is available as downloadable binaries from the [releases](https://github.com/raystack/compass/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`compass` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/raystack/compass/releases/latest) page:

```
scoop bucket add compass https://github.com/raystack/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update compass
```

#### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull raystack/compass:latest
```

To pull a specific version:

```
docker pull raystack/compass:v0.6.0
```

If you like to have a shell alias that runs the latest version of compass from docker whenever you type `compass`:

```
mkdir -p $HOME/.config/raystack
alias compass="docker run -e HOME=/tmp -v $HOME/.config/raystack:/tmp/.config/raystack --user $(id -u):$(id -g) --rm -it -p 8080:8080/tcp raystack/compass:latest"
```

## Usage

Compass provides a CLI, Connect RPC API (HTTP + gRPC), and an MCP server for AI agents.

#### CLI

Compass CLI is fully featured and simple to use. Run `compass --help` to see all available commands.

```
compass --help
```

Print command reference

```sh
compass reference
```

#### API

Compass provides a Connect RPC API that supports both Connect (HTTP) and gRPC protocols. Please refer to [proton](https://github.com/raystack/proton/tree/main/raystack/compass/v1beta1) for API definitions.

#### MCP Server

Compass exposes an MCP server at `/mcp` for AI agent integration. MCP-compatible systems can connect and use tools like `search_entities`, `get_context`, and `impact`.

## Contribute

Development of Compass happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and improvements.

Read compass [contribution guide](https://compass-raystack.vercel.app/docs/contribute/contribution.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Compass.

To help you get your feet wet and get you familiar with our contribution process, we have a list of [good first issues](https://github.com/raystack/compass/labels/good%20first%20issue) that contain bugs which have a relatively limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/raystack/compass/graphs/contributors).

## License

Compass is [Apache 2.0](LICENSE) licensed.
