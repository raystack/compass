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
-d, --data string       filter by field in asset.data
-o, --out -o json       flag to control output viewing, for json -o json (default "table")
    --page uint32       Number of pages
    --query string      querying by field
-s, --services string   filter by services
    --size uint32       Size of each page (default 20)
    --sort string       sort by certain fields
    --sort_dir string   sorting direction (asc / desc)
-t, --types string      filter by types
````

### `compass asset view <id>`

view asset for the given ID

## `compass configs`

Display configurations currently loaded

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

## `compass migrate`

Run storage migration

## `compass search <text> [flags]`

query the metadata available

```
-f, --filter string   --filter=field_key1:val1,key2:val2,key3:val3 gives exact match for values
-q, --query string    --query=--filter=field_key1:val1 supports fuzzy search
-r, --rankby string   --rankby=<numeric_field>
-s, --size uint32     --size=10 maximum size of response query
````

## `compass serve`

Serve gRPC & HTTP service

