# End to End Usage example

This document showcases columbus usage, from producing data, to using APIs for specific use-cases. This guide is geared towards adminstrators, more than users. It is meant to give the administrators an introduction of different configuration options available, and how they affect the behaviour of the application.

## Prerequisites

This guide assumes that you have a local instance of columbus running and listening on `localhost:8080`. See [Usage](usage.md) guide for information on how to run Columbus.

## Creating an type

Let’s say that you have a hypothetical tool called Piccolo and you have several deployments of this tool on your platform. Before we can push data for Piccolo deployments to Columbus, we need to first define the Piccolo type. To do this, we'll call to type create API `PUT /v1/types`

```text
$ curl -XPUT http://localhost:8080/v1/types/ \
  -H 'content-type: application/json' \
  -d '
{
    "name": "piccolo",
    "classification": "resource",
    "record_attributes": {
        "id": "deployment_name",
        "title": "name",
        "description": "desc",
        "labels": ["owner"]
    }
}'
```

This defines the type called ‘piccolo’. You can now push records for this type by making HTTP PUT calls to `/v1/types/piccolo/`. The record\_attributes field defines the partial schema for records that belong to this type.

## Adding Data

Now that we've defined our `piccolo` type, we can start pushing data for it.

```text
# notice the 'piccolo' in the API endpoint path
$ curl -XPUT http://localhost:8080/v1/types/piccolo/ \
  -H 'content-type: application/json' \
  -d'
[
    {
        "deployment_name": "01",
        "name": "deployment x 01",
        "desc": "this is the one",
        "owner": "jhon.doe",
        "created": "31/01/2019",
        "src": "sensu-01",
        "dest"; "gohan-01"
    },
    {
        "deployment_name": "02",
        "name": "deployment x 02",
        "desc": "this came second",
        "owner": "kami",
        "created": "24/02/2019"
        "src": "sensu-02",
        "dest"; "gohan-02"
    },
    {

        "deployment_name": "03",
        "name": "deployment x 03",
        "desc": "the last one",
        "owner": "kami",
        "created": "15/06/2017"
        "src": "sensu-03",
        "dest"; "gohan-03"
    }
]'
```

## Searching

Now we're ready to start searching. Let's run a search for the term 'one'

```text
$ curl http://localhost:8080/v1/search?text\=one | jq
[
  {
    "title": "deployment x 03",
    "id": "03",
    "type": "piccolo",
    "classification": "resource",
    "description": "the last one",
    "labels": {
      "owner": "kami"
    }
  },
  {
    "title": "deployment x 01",
    "id": "01",
    "type": "piccolo",
    "classification": "resource",
    "description": "this is the one",
    "labels": {
      "owner": "jhon.doe"
    }
  }
]
```

The search is run against ALL fields of the records. Notice how your data was transformed into the search results? This transformation is governed by the type definition. Most fields in type.record\_attributes map to one of the search result fields, apart from classification and type which come from the type definitions. labels can be any root level key in the record. Don’t worry if every record does not contain every label field. If any label field is missing from record, it’s just dropped from the search response for that record.

Search can be further restricted by specifying a filter criteria. For instance, if you wish to restrict the search to piccolo deployments that belong to `kami`, you can run:

```text
$ curl http://localhost:8080/v1/search?text=one&filter.owner=kami | jq
[
  {
    "title": "deployment x 03",
    "id": "03",
    "type": "piccolo",
    "classification": "resource",
    "description": "the last one",
    "labels": {
      "owner": "kami"
    }
  }
]
```

## Configuring Lineage

Now that we have configured the `piccolo` type and learnt how to use the search API to search it's records, let's configure lineage for it.

To begin with, let's add another type called `sensu` and add some records for it.

```text
$ curl -XPUT http://localhost:8080/v1/types/ \
  -H 'content-type: application/json' \
  -d '
{
    "name": "sensu",
    "classification": "dataset",
    "record_attributes": {
        "id": "id",
        "title": "name",
        "description": "desc",
        "labels": []
    }
}'

$ curl -XPUT http://localhost:8080/v1/types/sensu/ \
  -H 'content-type: application/json' \
  -d'
[
    {
        "id": "sensu-01",
        "name": "sensu-01",
        "desc": "primary sensu dataset"
    },
    {
        "id": "sensu-02",
        "name": "sensu-02",
        "desc": "secondary sensu dataset"
    },
    {
        "id": "sensu-03",
        "name": "sensu-03",
        "desc": "third sensu dataset"
    },
]'
```

`sensu` is the data store that `piccolo` instances read from. In order to configure lineage, we need to declare the lineage as a part of the type definition. In Columbus, any type can declare it's lineage, as long as it's records have information on the ID of related type instance.

For instance, if you look at the `src` field of `piccolo` instances, you'll see that they are id's of `sensu` instances. This means we can define the relationship between `piccolo` and `sensu` resources by declaring this relationship in `piccolo`'s definition. Note that it is sufficient \(and preferred\) that one type declare it's relationship to another. Both need not do this.

Let's update `piccolo` definition with the necessary lineage configuration.

```text
$ curl -XPUT http://localhost:8080/v1/types/ \
  -H 'content-type: application/json' \
  -d '
{
    "name": "piccolo",
    "classification": "resource",
    "record_attributes": {
        "id": "deployment_name",
        "title": "name",
        "description": "desc",
        "labels": ["owner"]
    },
    "lineage": [
        {
            "type": "sensu",
            "query": "$.src",
            "direction": "upstream"
        }
    ]
}'
```

The `lineage` declaration in the updated `piccolo` definition above can be read as: "each piccolo instance has an upstream resource of type "sensu", and the ID for that resource can be found using this JSON path query".

## Querying Lineage

To query lineage, we make a HTTP GET call to `/v1/lineage` API, specifying the type types that we're interested in.

```text
$ curl http://localhost:8080/v1/lineage?filter.type=piccolo
{
    "piccolo/deployment x 01": {
        "type": "piccolo",
        "urn": "deployment x 01",
        "upstreams": [
            "sensu/sensu-01"
        ],
        "downstreams": []
    },
    "piccolo/deployment x 02": {
        "type": "piccolo",
        "urn": "deployment x 02",
        "upstreams": [
            "sensu/sensu-02"
        ],
        "downstreams": []
    },
    "piccolo/deployment x 03": {
        "type": "piccolo",
        "urn": "deployment x 03",
        "upstreams": [
            "sensu/sensu-03"
        ],
        "downstreams": []
    },
    "sensu/sensu-01": {
        "type": "sensu",
        "urn": "sensu-01",
        "upstreams": [],
        "downstreams": [
            "piccolo/deployment x 01"
        ]
    },
    "sensu/sensu-02": {
        "type": "sensu",
        "urn": "sensu-02",
        "upstreams": [],
        "downstreams": [
            "piccolo/deployment x 02"
        ]
    },
    "sensu/sensu-03": {
        "type": "sensu",
        "urn": "sensu-03",
        "upstreams": [],
        "downstreams": [
            "piccolo/deployment x 03"
        ]
    }
}
```

The response represents a graph that consists of different type resources, and their dataflow relationships.

The lineage API also supports querying lineage for a single resource. For instance, to query the lineage of "sensu-01", we can run:

```text
$ curl http://localhost:8080/v1/lineage/sensu/sensu-01
{
    "piccolo/deployment x 01": {
        "type": "piccolo",
        "urn": "deployment x 01",
        "upstreams": [
            "sensu/sensu-01"
        ],
        "downstreams": []
    },
    "sensu/sensu-01": {
        "type": "sensu",
        "urn": "sensu-01",
        "upstreams": [],
        "downstreams": [
            "piccolo/deployment x 01"
        ]
    }
}
```

## Improving Relevancy

The last part of this guide demonstrates how to fine-tune the search itself. This is done via the "boost" configuration in the type definition. "boost" defines the "weight" of a specific field in each record. By default, all fields have a boost value of "1.0". We can change the "weight" of specific fields in documents to change their relevancy for search.

Explanation: when we run search, we look for the search terms in all the fields of each document. Then we score each hit depending on which fields match the query, and how well. Boost allows you to define the relative importance of each field within the document in context of search.

For instance, let's say that the `desc` field of `piccolo` records contain more important information than the rest of the fields, and that any search query that matches the `desc` should show up first in the results, rather than later.

To do this, we update the definition of `piccolo` with the updated boost configuration

```text
$ curl -XPUT http://localhost:8080/v1/types/ \
  -H 'content-type: application/json' \
  -d '
{
    "name": "piccolo",
    "classification": "resource",
    "record_attributes": {
        "id": "deployment_name",
        "title": "name",
        "description": "desc",
        "labels": ["owner"]
    },
    "lineage": [
        {
            "type": "sensu",
            "query": "$.src",
            "direction": "upstream"
        }
    ],
    "boost": {
        "desc": 1.5
    }
}'
```

Now, any search query that matches the value of `desc` will result in a higher score for that document than if it had matched say the `deployment_name` field.

