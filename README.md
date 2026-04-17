# Compass

![test workflow](https://github.com/raystack/compass/actions/workflows/test.yml/badge.svg)
![release workflow](https://github.com/raystack/compass/actions/workflows/release.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/raystack/compass/badge.svg?branch=main)](https://coveralls.io/github/raystack/compass?branch=main)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/raystack/compass?logo=semantic-release)](Version)

Compass is a context engine that builds a temporal knowledge graph of your organization — entities, relationships, and documents across every system — queryable by humans and AI agents alike. Search it. Traverse it. Reason over it.

## Key Features

- **AI-Native** -- Composable MCP tools let AI agents search, traverse, assess impact, and read documents over your organizational graph.
- **Knowledge Graph** -- Typed, directed, temporal relationships between entities including lineage, ownership, and custom edge types.
- **Context Assembly** -- Multi-hop bidirectional traversal builds a subgraph around any entity — the full picture in one call.
- **Impact Analysis** -- Blast radius analysis traces what depends on what and what breaks when something changes.
- **Hybrid Search** -- Keyword precision with semantic similarity using Postgres-native full-text search and pgvector embeddings.
- **Documents** -- Attach any document to entities — runbooks, decisions, postmortems, design docs — all indexed for semantic search.
- **Open Type System** -- Any entity type, any edge type, any properties. Hardware, software, operations, people — anything fits.

## Documentation

- [Quickstart](https://compass-raystack.vercel.app/docs/quickstart) -- Get running in 5 minutes
- [Guides](https://compass-raystack.vercel.app/docs/guides/entities) -- Entities, edges, search, context, MCP, CLI, API
- [Internals](https://compass-raystack.vercel.app/docs/internals/architecture) -- Architecture, search engine, storage

## Installation

Install Compass on macOS, Windows, Linux, or via Docker.

#### macOS

```sh
brew install raystack/tap/compass
```

#### Linux

Download `.deb` or `.rpm` from [releases](https://github.com/raystack/compass/releases/latest):

```sh
sudo dpkg -i compass_*.deb
```

#### Docker

```sh
docker pull raystack/compass:latest
```

#### Build from Source

```sh
git clone https://github.com/raystack/compass.git
cd compass && make
```

## Usage

```bash
# Start PostgreSQL
docker-compose up -d

# Initialize and run
compass config init
compass server migrate
compass server start

# Search the graph
compass entity search "orders" --mode hybrid

# Explore context
compass entity context urn:bigquery:orders --depth 2

# Analyze impact
compass entity impact urn:kafka:events --depth 3
```

### MCP Server

Connect AI agents to Compass via MCP. Add to `.mcp.json`:

```json
{
  "mcpServers": {
    "compass": {
      "type": "sse",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## License

Compass is [Apache 2.0](LICENSE) licensed.
