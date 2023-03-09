# CLI

## `compass asset`

Manage assets

### `compass asset delete <id>`

delete asset with the given ID

### `compass asset edit [flags]`

upsert a new asset or patch

```
-b, --body string   filepath to body that has to be upserted
````

### `compass asset list [flags]`

lists all assets

```
-d, --data stringToString   filter by field in asset.data (default [])
-o, --out -o json           flag to control output viewing, for json -o json (default "table")
    --page uint32           Page number offset (starts from 0)
    --query string          querying by field
    --query_fields string   querying by fields
-s, --services string       filter by services
    --size uint32           Size of each page (default 10)
    --sort string           sort by certain fields
    --sort_dir string       sorting direction (asc / desc)
-t, --types string          filter by types
````

### `compass asset star <id>`

star an asset by id for current user

### `compass asset stargazers <id> [flags]`

list all stargazers for a given asset id

```
--page uint32   Page number offset (starts from 0)
--size uint32   Size of each page (default 10)
````

### `compass asset starred [flags]`

list all the starred assets for current user

```
-o, --out -o json   flag to control output viewing, for json -o json (default "table")
    --page uint32   Page number offset (starts from 0)
    --size uint32   Size of each page (default 10)
````

### `compass asset types [flags]`

lists all asset types

```
-d, --data stringToString   filter by field in asset.data (default [])
    --query string          filter by specific query
    --query_fields string   filter by query field
-s, --services string       filter by services
-t, --types string          filter by types
````

### `compass asset unstar <id>`

unstar an asset by id for current user

### `compass asset version <urn> <version>`

get asset's previous version by urn or id and version number

### `compass asset versionhistory <id> [flags]`

get asset version history by id

```
--page uint32   Page number offset (start from 0)
--size uint32   Size of each page (default 10)
````

### `compass asset view <urn>`

view asset for the given ID or URN

## `compass completion [bash|zsh|fish|powershell]`

Generate shell completion scripts

## `compass config <command>`

Manage server and client configurations

### `compass config init`

Initialize a new sevrer and client configuration

### `compass config list`

List server and client configuration settings

## `compass discussion`

Manage discussions

### `compass discussion list [flags]`

lists all discussions

```
-o, --out -o json   flag to control output viewing, for json -o json (default "table")
````

### `compass discussion post [flags]`

post discussions, add 

```
-b, --body string   filepath to body that has to be upserted
````

### `compass discussion view <id>`

view discussion for the given ID

## `compass lineage <urn>`

observe the lineage of metadata

## `compass search <text> [flags]`

query the metadata available

```
-f, --filter string   --filter=field_key1:val1,key2:val2,key3:val3 gives exact match for values
-q, --query string    --query=--filter=field_key1:val1 supports fuzzy search
-r, --rankby string   --rankby=<numeric_field>
-s, --size uint32     --size=10 maximum size of response query
````

## `compass server <command>`

Run compass server

### `compass server migrate`

Run storage migration

### `compass server start`

Start server on default port 8080

## `compass version`

Print version information

