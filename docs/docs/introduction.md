---
id: introduction
slug: /
---

# Introduction

Welcome to the introductory guide to Compass! This guide is the best place to start with Compass. We cover what Compass is, what problems it can solve, how it works, and how you can get started using it. If you are familiar with the basics of Compass, the guide provides a more detailed reference of available features.

## What is Compass?

Compass is a context engine that builds a knowledge graph of your organization's metadata, capturing entities, relationships, and lineage across systems and time, making it discoverable and queryable for both humans and AI agents.

Critical organizational knowledge lives scattered across dozens of systems: services, datasets, applications, teams, configurations, decisions, and the relationships between them. Compass resolves observations from these sources into unified entities, constructs a temporal graph of their relationships, and indexes everything for both keyword and semantic search. The result is a context graph that stitches together what exists, who owns it, how it connects, and what changed over time, so both humans and AI agents can discover, traverse, and reason over the full picture.

![](/assets/overview.svg)

## The Problem

Organizational knowledge is fragmented. The same logical entity appears across multiple systems with different names, schemas, and levels of detail. Relationships between entities live in people's heads, scattered across wikis, chat threads, and tribal knowledge. When humans need context, they spend hours stitching it together manually. When AI agents need context, they have nowhere to look.

This fragmentation compounds as organizations grow. Teams cannot find what already exists. Dependencies are invisible until something breaks. Ownership is unclear. Decisions are made without the full picture because assembling that picture takes too long.

Compass solves this by acting as the resolution and serving layer for organizational metadata. It takes raw observations from collection systems like Meteor, resolves them into unified entities, builds a graph of their relationships, and makes everything searchable and traversable through APIs that serve both human interfaces and AI agents.

## Key Features

- **Entity Resolution:** Resolve and deduplicate metadata observations from multiple sources into unified entities with stable identity.
- **Knowledge Graph:** Store typed, directed relationships between entities including lineage, ownership, documentation, and custom edge types.
- **Hybrid Search:** Combine keyword precision with semantic similarity using Postgres-native full-text search and pgvector embeddings.
- **Graph Traversal:** Multi-hop traversal queries across the entity graph for impact analysis, dependency tracking, and path discovery.
- **Context Composition:** Assemble schema, lineage, ownership, and quality signals into context documents ready for LLM consumption.
- **AI Serving:** Expose the full graph as an MCP server so AI agents can discover, traverse, and reason over organizational knowledge.
- **Extensibility:** Open type system for entities and relationships to support any kind of metadata across your infrastructure.

## Using Compass

You can interact with Compass in any of the following ways:

### Command Line Interface

You can use the Compass command line interface to issue commands and manage the server. Using the command line can be faster and more convenient than the console. For more information on using the Compass CLI, see the [CLI Reference](./reference/cli.md) page.

### HTTP and gRPC APIs

Compass provides HTTP and gRPC APIs for programmatic access. The API is built with [Connect RPC](https://connectrpc.com/) and supports both Connect and gRPC protocols. For more information, see the [API reference](./reference/api.md) page.

### MCP Server

Compass exposes itself as an MCP (Model Context Protocol) server, allowing any MCP-compatible AI system to connect and use tools like search, lineage traversal, schema lookup, and context assembly.

## Where to go from here

See the [installation](./installation) page to install the Compass CLI. Next, we recommend completing the guides. The tour provides an overview of most of the existing functionality of Compass and takes approximately 20 minutes to complete.

After completing the tour, check out the remainder of the documentation in the reference and concepts sections for your specific areas of interest. We've aimed to provide as much documentation as we can for the various components of Compass to give you a full understanding of Compass's surface area.

Finally, follow the project on [GitHub](https://github.com/raystack/compass), and contact us if you'd like to get involved.
