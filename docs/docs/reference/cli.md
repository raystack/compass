# CLI

Compass CLI provides commands to manage the server, entities, namespaces, and configuration. Run `compass --help` to see all available commands.

## Global Flags

```
-c, --config string   Override config file
```

## `compass config <command>`

Manage server and client configurations

### `compass config init`

Initialize a new server and client configuration

### `compass config list`

List server and client configuration settings

## `compass entity`

Manage entities in the knowledge graph. Alias: `entities`

### `compass entity list [flags]`

List all entities

```
    --types string    filter by types (comma-separated)
    --source string   filter by source
-q, --query string    search query
    --size uint32     page size (default 20)
    --offset uint32   page offset (default 0)
```

### `compass entity view <id>`

View entity details by ID or URN

### `compass entity upsert [flags]`

Create or update an entity

```
    --urn string           entity URN (required)
    --type string          entity type (required)
    --name string          entity name (required)
    --description string   description
    --source string        source system
```

### `compass entity delete <urn>`

Delete an entity by URN

### `compass entity search <text> [flags]`

Search entities using keyword, semantic, or hybrid mode

```
    --types string    filter by types
    --source string   filter by source
    --mode string     search mode: keyword, semantic, hybrid (default "keyword")
    --size uint32     max results (default 10)
```

### `compass entity types`

List all entity types with counts

### `compass entity context <urn> [flags]`

Get full context subgraph for an entity

```
    --depth uint32   traversal depth (default 2)
```

### `compass entity impact <urn> [flags]`

Analyze downstream blast radius for an entity

```
    --depth uint32   traversal depth (default 3)
```

## `compass namespace`

Manage namespaces. Alias: `ns`

### `compass namespace create [flags]`

Create a new namespace

```
-n, --name string    namespace unique name
-s, --state string   is namespace shared with existing tenants or a dedicated one (default "shared")
```

### `compass namespace list [flags]`

List all namespaces

```
-o, --out -o json   flag to control output viewing, for json -o json (default "table")
```

### `compass namespace view <id>`

View namespace for the given uuid or name

## `compass server <command>`

Run compass server. Alias: `s`

### `compass server migrate [flags]`

Run storage migration

```
--down   rollback migration one step
```

### `compass server start`

Start server on default port 8080

## `compass version`

Print version information
