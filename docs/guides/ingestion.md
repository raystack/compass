# Ingesting metadata

This document showcases compass usage, from producing data, to using APIs for specific use-cases. This guide is geared towards adminstrators, more than users. It is meant to give the administrators an introduction of different configuration options available, and how they affect the behaviour of the application.

## Prerequisites

This guide assumes that you have a local instance of compass running and listening on `localhost:8080`. See [Installation](installation.md) guide for information on how to run Compass.

## Adding Data

Let’s say that you have a hypothetical tool called Piccolo and you have several deployments of this tool on your platform. Before we can push data for Piccolo deployments to Compass, you need to recognize the type of Piccolo, whether it is a kind of `table`, `topic`, `dashboard`, or `job`. One can ingest metadata to compass with the Upsert Patch API. The API contract is available [here](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

If there is an existing asset, Upsert Patch API will check each field whether there is an update in the field of the existing asset. With this behaviour, it is possible to send partial updated field to update a certain field only as long as the `urn`, `type`, and `service` match with the existing asset. If there is any field changed, a new version of the asset will be created. If the asset does not exist, upsert patch API will create a new asset. Apart from asset details, we also could send upstreams and downstreams of lineage edges of the asset in the body.

Let's say `piccolo` tool is a kind of `table`, we can start pushing data for it. Let's add 3 metadata of `picollo`.

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
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

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
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

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
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

## Searching

Now we're ready to start searching. Let's run a search for the term 'one'

```bash
$ curl 'http://localhost:8080/v1beta1/search?text\=one' \
--header 'Compass-User-UUID:odpf@email.com'  | jq

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

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=one&query[owners]=kami' \
--header 'Compass-User-UUID:odpf@email.com' | jq

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

## Lineage

Now that we have configured the `piccolo` type and learnt how to use the search API to search it's assets, let's configure lineage for it.

To begin with, let's start over adding picolo metadata with its lineage information and add another metadata with service name `sensu` and type `topic` and add some records for it.

### Adding `picollo` Metadata
```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
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
            "urn": sensu:deployment-01",
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

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
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
            "urn": sensu:deployment-02",
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

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
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
            "urn": sensu:deployment-03",
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
### Adding `sensu` Metadata
```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
--data-raw '{
    "asset": {
        "urn": "sensu:deployment-01",
        "type": "topic",
        "name": "deployment-01",
        "service": "sensu",
        "description": "primary sensu dataset"
    },
    "upstreams": [],
    "downstreams": []
}'

$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
--data-raw '{
    "asset": {
        "urn": "sensu:deployment-02",
        "type": "topic",
        "name": "deployment-02",
        "service": "sensu",
        "description": "secondary sensu dataset"
    },
    "upstreams": [],
    "downstreams": []
}'

$ curl --request PATCH 'http://localhost:8080/v1beta1/assets' \
--header 'Compass-User-UUID:odpf@email.com' \
--data-raw '{
    "asset": {
        "urn": "sensu:deployment-02",
        "type": "topic",
        "name": "deployment-02",
        "service": "sensu",
        "description": "tertiary sensu dataset"
    },
    "upstreams": [],
    "downstreams": []
}'
```

`sensu` is the data store that `piccolo` instances read from. In order to have a lineage, we need to have the metadata urn of `sensu` in Compass.

For instance, if you look at the `upstreams` and `downstreams` fields when we are ingesting `piccolo` metadata, you'll see that they are urn's of `sensu` instances. This means we can define the relationship between `piccolo` and `sensu` resources by declaring this relationship in `piccolo`'s definition. Note that it is sufficient \(and preferred\) that one declare it's relationship to another. Both need not do this.

### Querying Lineage

To query lineage, we make a HTTP GET call to `/v1beta1/lineage` API, specifying the metadata that we're interested in.

```bash
curl 'http://localhost:8080/v1beta1/lineage/picollo%3Adeployment-01' \
--header 'Compass-User-UUID:odpf@email.com'

{
    data: [
        {
            "source": {
                "urn": "picollo:deployment-01",
                "type": "table",
                "service": "picollo",
            },
            "target": {
                "urn": "gohan:deployment-01",
                "type": "table",
                "service": "gohan",
            },
            "props": nil
        },
        {
            "source": {
                "urn": "sensu:deployment-01",
                "type": "topic",
                "service": "sensu",
            },
            "target": {
                "urn": "picollo:deployment-01",
                "type": "table",
                "service": "picollo",
            },
            "props": nil
        }
    ]
}
```

The response represents a graph that consists of edges in its graph.
