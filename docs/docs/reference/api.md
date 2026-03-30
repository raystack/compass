# API Reference

Compass provides a Connect RPC API that supports both Connect (HTTP) and gRPC protocols. All endpoints are under the `raystack.compass.v1beta1.CompassService` service.

API definitions are maintained in [raystack/proton](https://github.com/raystack/proton/tree/main/raystack/compass/v1beta1).

**License:** [Apache License 2.0](https://github.com/raystack/compass/blob/main/LICENSE)

## Endpoints

### Entity

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| GET | `GetAllEntities` | List all entities, optionally filtered by types, source, or query |
| GET | `GetEntityByID` | Get a single entity by ID or URN |
| POST | `UpsertEntity` | Create or update an entity |
| DELETE | `DeleteEntity` | Delete an entity by URN |
| GET | `SearchEntities` | Search entities using keyword, semantic, or hybrid mode |
| GET | `SuggestEntities` | Get entity name suggestions for autocomplete |
| GET | `GetEntityTypes` | List all entity types with counts |

### Entity Context & Impact

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| GET | `GetEntityContext` | Get full context subgraph for an entity (entity, relationships, related entities) |
| GET | `GetEntityImpact` | Analyze downstream blast radius for an entity |

### Edge

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| POST | `UpsertEdge` | Create or update a typed, directed edge between two entities |
| GET | `GetEdges` | Get edges for an entity, optionally filtered by type and direction |
| DELETE | `DeleteEdge` | Delete an edge by ID |

### Star

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| POST | `StarEntity` | Star an entity for a user |
| DELETE | `UnstarEntity` | Unstar an entity for a user |
| GET | `GetUserStarredEntities` | List starred entities for a specific user |
| GET | `GetMyStarredEntities` | List starred entities for the current user |
| GET | `GetMyStarredEntity` | Check if the current user has starred a specific entity |
| GET | `GetEntityStargazers` | List users who have starred an entity |

### Namespace

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| POST | `CreateNamespace` | Create a new namespace |
| GET | `GetNamespace` | Get a namespace by ID or name |
| PATCH | `UpdateNamespace` | Update a namespace |
| GET | `ListNamespaces` | List all namespaces |

## MCP Server

Compass exposes an MCP (Model Context Protocol) server at the `/mcp` endpoint. Any MCP-compatible AI system can connect and use the following tools:

| Tool | Parameters | Description |
| ---- | ---------- | ----------- |
| `search_entities` | `text` (required), `types`, `source`, `mode`, `size` | Search the entity knowledge graph with keyword, semantic, or hybrid mode |
| `get_context` | `urn` (required), `depth` | Get full context about an entity including relationships and related entities |
| `impact` | `urn` (required), `depth` | Analyze downstream blast radius for an entity |

## Health Check

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | `/ping` | Health check endpoint, returns "pong" |

## Authentication

Compass requires an identity header in all API requests. The header key is configurable, with the default being `Compass-User-UUID`. An optional email header (`Compass-User-Email`) can also be provided.

Namespace isolation is controlled via the `x-namespace` header. If not provided, requests fall back to the `default` namespace. Namespace can also be passed as a `namespace_id` claim in a JWT bearer token.

## Protocol

Compass uses [Connect RPC](https://connectrpc.com/), which means you can call the API using:

- **Connect protocol** (HTTP POST with JSON or Protobuf)
- **gRPC protocol** (HTTP/2 with Protobuf)
- **gRPC-Web protocol** (for browser clients)

Example using curl with Connect protocol:

```bash
curl \
  --header "Content-Type: application/json" \
  --header "Compass-User-UUID: user@example.com" \
  --data '{"text": "revenue", "mode": "hybrid"}' \
  http://localhost:8080/raystack.compass.v1beta1.CompassService/SearchEntities
```
