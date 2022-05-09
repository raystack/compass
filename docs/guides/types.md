# Ingesting metadata

This document showcases compass usage, from producing data, to using APIs for specific use-cases. This guide is geared towards adminstrators, more than users. It is meant to give the administrators an introduction of different configuration options available, and how they affect the behaviour of the application.


## Prerequisites

This guide assumes that you have a local instance of compass running and listening on `localhost:8080`. See [Usage](usage.md) guide for information on how to run Compass.

## Adding Data

Let’s say that you have a hypothetical tool called Piccolo and you have several deployments of this tool on your platform. Before we can push data for Piccolo deployments to Compass, you need to recognize the type of Piccolo, whether it is a kind of `table`, `topic`, `dashboard`, or `job`.

Let's say `piccolo` tool is a kind of `table`, we can start pushing data for it. Let's add 3 metadata of `picollo`.


```text
$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
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

```text
$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
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

```text
$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
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

## Searching

Now we're ready to start searching. Let's run a search for the term 'one'

```text
$ curl http://localhost:8080/v1beta1/search?text\=one --header 'Compass-User-UUID:odpf@email.com'  | jq
```text
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

```text
$ curl http://localhost:8080/v1/search?text=one&query[owners]=kami | jq
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

## Querying Lineage

Now that we have configured the `piccolo` type and learnt how to use the search API to search it's records, let's configure lineage for it.

To begin with, let's add another metadata with service name `sensu` and type `topic` and add some records for it.

```text
$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
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

$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
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

$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
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


To query lineage, we make a HTTP GET call to `/v1beta1/lineage` API, specifying the metadata that we're interested in.

```text
$ curl http://localhost:8080/v1beta1/lineage/picollo%3Adeployment-01 --header 'Compass-User-UUID:odpf@email.com'
{
    data: [
        {
            "source": {
                "urn": picollo:deployment-01",
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
                "urn": sensu:deployment-01",
                "type": "topic",
                "service": "sensu",
            },
            "target": {
                "urn": picollo:deployment-01",
                "type": "table",
                "service": "picollo",
            },
            "props": nil
        },
    ]
}
```

The response represents a graph that consists of edges in its graph.
