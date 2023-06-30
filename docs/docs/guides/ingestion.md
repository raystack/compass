import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Ingesting metadata

This document showcases compass usage, from producing data, to using APIs for specific use-cases. This guide is geared towards adminstrators, more than users. It is meant to give the administrators an introduction of different configuration options available, and how they affect the behaviour of the application.

## Prerequisites

This guide assumes that you have a local instance of compass running and listening on `localhost:8080`. See [Installation](installation.md) guide for information on how to run Compass.

## Adding Data

Let’s say that you have a hypothetical tool called Piccolo and you have several deployments of this tool on your platform. Before we can push data for Piccolo deployments to Compass, you need to recognize the type of Piccolo, whether it is a kind of `table`, `topic`, `dashboard`, or `job`. One can ingest metadata to compass with the Upsert Patch API. The API contract is available [here](https://github.com/raystack/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

If there is an existing asset, Upsert Patch API will check each field whether there is an update in the field of the existing asset. With this behaviour, it is possible to send partial updated field to update a certain field only as long as the `urn`, `type`, and `service` match with the existing asset. If there is any field changed, a new version of the asset will be created. If the asset does not exist, upsert patch API will create a new asset. Apart from asset details, we also could send upstreams and downstreams of lineage edges of the asset in the body.

Let's say `piccolo` tool is a kind of `table`, we can start pushing data for it. Let's add 3 metadata of `picollo`.

<Tabs>
<TabItem value="" label="Picollo 1">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "picollo:deployment-01",
        "type": "table",
        "name": "deployment-01",
        "service": "picollo",
        "description": "this is the one",
        "data": {},
        "owners": [
            {
                "email": "john.doe@email.com"
            }
        ]
    }
}'
```

</TabItem>
<TabItem value="picollo 2" label="Picollo 2">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "picollo:deployment-02",
        "type": "table",
        "name": "deployment-02",
        "service": "picollo",
        "description": "this came second",
        "data": {},
        "owners": [
            {
                "email": "kami@email.com"
            }
        ]
    }
}'
```

</TabItem>
<TabItem value="picollo 3" label="Picollo 3">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "picollo:deployment-03",
        "type": "table",
        "name": "deployment-03",
        "service": "picollo",
        "description": "the last one",
        "data": {},
        "owners": [
            {
                "email": "kami@email.com"
            }
        ]
    }
}'
```

</TabItem>
</Tabs>

## Searching

#### We can search for required text in the following ways:

1. Using **`compass search <text>`** CLI command
2. Calling to **`GET /v1beta1/search`** API with `text` to be searched as query parameter

Now we're ready to start searching. Let's run a search for the term **'one'** from the assets we ingested earlier.

<Tabs groupId="cli">
<TabItem value="CLI" label="CLI">

```bash
$ compass search one
```

</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text\=one' \
--header 'Compass-User-UUID:raystack@email.com'  | jq
```

</TabItem>
</Tabs>

The output is the following:

```json
{
  "data": [
    {
      "urn": "picollo:deployment-01",
      "type": "table",
      "name": "deployment-01",
      "service": "picollo",
      "description": "this is the one",
      "data": {},
      "owners": [
        {
          "email": "john.doe@email.com"
        }
      ],
      "labels": {}
    },
    {
      "urn": "picollo:deployment-03",
      "type": "table",
      "name": "deployment-03",
      "service": "picollo",
      "description": "the last one",
      "data": {},
      "owners": [
        {
          "email": "kami@email.com"
        }
      ],
      "labels": {}
    }
  ]
}
```

The search is run against ALL fields of the records. It can be further restricted by specifying a filter criteria, could be exact match with `filter` and fuzzy match with `query`. For instance, if you wish to restrict the search to piccolo deployments that belong to `kami` (fuzzy), you can run:

#### We can search with custom queries in the following ways:

1. Using **`compass search <text> --query=field_key1:val1`** CLI command
2. Calling to **`GET /v1beta1/search`** API with `text` and `query[field_key1]=val1` as query parameters

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search one --query=owners:kami
```

</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=one&query[owners]=kami' \
--header 'Compass-User-UUID:raystack@email.com' | jq
```

</TabItem>
</Tabs>

The output is the following:

```json
{
  "data": [
    {
      "urn": "picollo:deployment-03",
      "type": "table",
      "name": "deployment-03",
      "service": "picollo",
      "description": "the last one",
      "data": {},
      "owners": [
        {
          "email": "kami@email.com"
        }
      ],
      "labels": {}
    }
  ]
}
```

## Lineage

Now that we have configured the `piccolo` type and learnt how to use the search API to search it's assets, let's configure lineage for it.

To begin with, let's start over adding picolo metadata with its lineage information and add another metadata with service name `sensu` and type `topic` and add some records for it.

### Adding `picollo` Metadata

<Tabs>
<TabItem value="picollo 1" label="Picollo 1">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "picollo:deployment-01",
        "type": "table",
        "name": "deployment-01",
        "service": "picollo",
        "description": "this is the one",
        "data": {},
        "owners": [
            {
                "email": "john.doe@email.com"
            }
        ]
    },
    "upstreams": [
        {
            "urn": "sensu:deployment-01",
            "type": "topic",
            "service": "sensu"
        }
    ],
    "downstreams": [
        {
            "urn": "gohan:deployment-01",
            "type": "table",
            "service": "gohan"
        }
    ]
}'
```

</TabItem>
<TabItem value="picollo 2" label="Picollo 2">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "picollo:deployment-02",
        "type": "table",
        "name": "deployment-02",
        "service": "picollo",
        "description": "this came second",
        "data": {},
        "owners": [
            {
                "email": "kami@email.com"
            }
        ]
    },
    "upstreams": [
        {
            "urn": "sensu:deployment-02",
            "type": "topic",
            "service": "sensu"
        }
    ],
    "downstreams": [
        {
            "urn": "gohan:deployment-02",
            "type": "table",
            "service": "gohan"
        }
    ]
}'
```

</TabItem>
<TabItem value="picollo 3" label="Picollo 3">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "picollo:deployment-03",
        "type": "table",
        "name": "deployment-03",
        "service": "picollo",
        "description": "the last one",
        "data": {},
        "owners": [
            {
                "email": "kami@email.com"
            }
        ]
    },
    "upstreams": [
        {
            "urn": "sensu:deployment-03",
            "type": "topic",
            "service": "sensu"
        }
    ],
    "downstreams": [
        {
            "urn": "gohan:deployment-03",
            "type": "table",
            "service": "gohan"
        }
    ]
}'
```

</TabItem>
</Tabs>

### Adding `sensu` Metadata

`sensu` is the data store that `piccolo` instances read from. In order to have a lineage, we need to have the metadata urn of `sensu` in Compass.

For instance, if you look at the `upstreams` and `downstreams` fields when we are ingesting `piccolo` metadata, you'll see that they are urn's of `sensu` instances. This means we can define the relationship between `piccolo` and `sensu` resources by declaring this relationship in `piccolo`'s definition.
<Tabs>
<TabItem value="sensu 1" label="Sensu 1">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "sensu:deployment-01",
        "type": "topic",
        "name": "deployment-01",
        "service": "sensu",
        "description": "primary sensu dataset",
        "data": {}
    },
    "upstreams": [],
    "downstreams": []
}'
```

</TabItem>
<TabItem value="sensu 2" label="Sensu 2">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "sensu:deployment-02",
        "type": "topic",
        "name": "deployment-02",
        "service": "sensu",
        "description": "secondary sensu dataset",
        "data": {}
    },
    "upstreams": [],
    "downstreams": []
}'
```

</TabItem>
<TabItem value="sensu 3" label="Sensu 3">

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:raystack@email.com' \
--data-raw '{
    "asset": {
        "urn": "sensu:deployment-03",
        "type": "topic",
        "name": "deployment-03",
        "service": "sensu",
        "description": "tertiary sensu dataset",
        "data": {}
    },
    "upstreams": [],
    "downstreams": []
}'
```

</TabItem>
</Tabs>

**Note:** it is sufficient \(and preferred\) that one declare it's relationship to another. Both need not do this.

### Querying Lineage

#### We can search for lineage in the following ways:

1. Using **`compass lineage <urn>`** CLI command
2. Calling to **`GET /v1beta1/lineage/:urn`** API with `urn` to be searched as the path parameter

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass lineage picollo:deployment-01
```

</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
curl 'http://localhost:8080/v1beta1/lineage/picollo%3Adeployment-01' \
--header 'Compass-User-UUID:raystack@email.com'
```

</TabItem>
</Tabs>

The output is the following:

```json
{
  "data": [
    {
      "source": "sensu:deployment-01",
      "target": "picollo:deployment-01",
      "prop": {
        "root": "picollo:deployment-01"
      }
    },
    {
      "source": "picollo:deployment-01",
      "target": "gohan:deployment-01",
      "prop": {
        "root": "picollo:deployment-01"
      }
    }
  ],
  "node_attrs": {}
}
```

The response represents a graph that consists of edges in its graph.
