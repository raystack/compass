# Usage

## Prerequisites

This guide assumes that you have a local instance of compass running and listening on `localhost:8080`. See [Installation](installation.md) guide for information on how to run Compass.

## Storage Migration
Compass has a `migrate` command that could be used to initialize the system. The command will migrate PostgreSQl and Elasticsearch.

## Required Header/Metadata in API
Compass has a concept of [User](../concepts/user.md). In the current version, all HTTP & gRPC APIs in Compass requires an identity header/metadata in the request. The header key is configurable but the default name is `Compass-User-UUID`.

Compass APIs also expect an additional optional e-mail header. This is also configurable and the default name is `Compass-User-Email`. The purpose of having this optional e-mail header is described in the [User](../concepts/user.md) section.

## Using the Search API

The API contract is available [here](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

To demonstrate how to use compass, we’re going to query it for resources that contain the word ‘booking’.

```text
$ curl http://localhost:8080/v1beta1/search?text=booking --header 'Compass-User-UUID:odpf@email.com' 
```

This will return a list of search results. Here’s a sample response:

```text
{
    "data": [
        {
            "id": "00c06ef7-badb-4236-9d9e-889697cbda46",
            "urn": "kafka::g-godata-id-playground/ g-godata-id-seg-enriched-booking-dagger",
            "type": "topic",
            "service": "kafka",
            "name": "g-godata-id-seg-enriched-booking-dagger",
            "description": "",
            "labels": {
                "flink_name": "g-godata-id-playground",
                "sink_type": "kafka"
            }
        },
        {
            "id": "9e69c08a-c3c2-4e04-957f-c8010c1e6515",
            "urn": "kafka::g-godata-id-playground/ g-godata-id-booking-bach-test-dagger",
            "type": "topic",
            "service": "kafka",
            "name": "g-godata-id-booking-bach-test-dagger",
            "description": "",
            "labels": {
                "flink_name": "g-godata-id-playground",
                "sink_type": "kafka"
            }
        },
        {
            "id": "ff597a0f-8062-4370-a54c-fd6f6c12d2a0",
            "urn": "kafka::g-godata-id-playground/ g-godata-id-booking-bach-test-3-dagger",
            "type": "topic",
            "service": "kafka",
            "title": "g-godata-id-booking-bach-test-3-dagger",
            "description": "",
            "labels": {
                "flink_name": "g-godata-id-playground",
                "sink_type": "kafka"
            }
        }
    ]
}
```

Compass decouple identifier from external system with the one that is being used internally. ID is the internally auto-generated unique identifier. URN is the external identifier of the asset, while Name is the human friendly name for it. See the complete API spec to learn more about what the rest of the fields mean.

### Filter
Compass search API also supports restricting search results via filter by passing it in query params.
Filter query params format is `filter[{field_key}]={value}` where `field_key` is the field name that we want to restrict and `value` is what value that should be matched. Filter could also support nested field by chaining key `field_key` with `.` \(dot\) such as `filter[{field_key}.{nested_field_key}]={value}`. For instance, to restrict search results to the ‘id’ landscape for ‘odpf’ organisation, run:

$ curl [http://localhost:8080/v1beta1/search?text=booking&filter[labels.landscape]=vn&filter[labels.entity]=odpf](http://localhost:8080/v1beta1/search?text=booking&filter[labels.landscape]=vn&filter[labels.entity]=odpf) --header 'Compass-User-UUID:odpf@email.com' 

Under the hood, filter's work by checking whether the matching document's contain the filter key and checking if their values match. Filters can be specified multiple times to specify a set of filter criteria. For example, to search for ‘booking’ in both ‘vn’ and ‘th’ landscape, run:

```text
$ curl http://localhost:8080/v1beta1/search?text=booking&filter[labels.landscape]=id&filter[labels.landscape]=th --header 'Compass-User-UUID:odpf@email.com' 
```

### Query
Apart from filters, Compass search API also supports fuzzy restriction in its query params. The difference of filter and query are, filter is for exact match on a specific field in asset while query is for fuzzy match.

Query format is not different with filter `query[{field_key}]={value}` where `field_key` is the field name that we want to query and `value` is what value that should be fuzzy matched. Query could also support nested field by chaining key `field_key` with `.` \(dot\) such as `query[{field_key}.{nested_field_key}]={value}`. For instance, to search results that has a name `kafka` and belongs to the team `data_engineering`, run:

```text
$ curl http://localhost:8080/v1beta1/search?text=booking&query[name]=kafka&query[labels.team]=data_eng --header 'Compass-User-UUID:odpf@email.com' 
```

### Ranking Results
Compass allows user to rank the results based on a numeric field in the asset. It supports nested field by using the `.` \(dot\) to point to the nested field. For instance, to rank the search results based on `usage_count` in `data` field, run:

```text
$ curl http://localhost:8080/v1beta1/search?text=booking&rankby=data.usage_count --header 'Compass-User-UUID:odpf@email.com' 
```

### Size
You can also specify the number of maximum results you want compass to return using the ‘size’ parameter

```text
$ curl http://localhost:8080/v1beta1/search?text=booking&size=5 --header 'Compass-User-UUID:odpf@email.com' 
```

## Using the Suggest API
The Suggest API gives a number of suggestion based on asset's name. There are 5 suggestions by default return by this API.

The API contract is available [here](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

Example of searching assets suggestion that has a name ‘booking’.

```text
$ curl http://localhost:8080/v1beta1/search/suggest?text=booking --header 'Compass-User-UUID:odpf@email.com' 
```
This will return a list of suggestions. Here’s a sample response:

```text
{
    "data": [
        "booking-daily-test-962ZFY",
        "booking-daily-test-c7OUZv",
        "booking-weekly-test-fmDeUf",
        "booking-daily-test-jkQS2b",
        "booking-daily-test-m6Oe9M"
    ]
}
```
## Using the Get Assets API
The Get Assets API returns assets from Compass' main storage (PostgreSQL) while the Search API returns assets from Elasticsearch. The Get Assets API has several options (filters, size, offset, etc...) in its query params.


|  Query Params | Description |
|---|---|
|`types=topic,table`| filter by types |
|`services=kafka,postgres`| filter by services |
|`data[dataset]=booking&data[project]=p-godata-id`| filter by field in asset.data |
|`q=internal&q_fields=name,urn,description,services`| querying by field|
|`sort=created_at`|sort by certain fields|
|`direction=desc`|sorting direction (asc / desc)|


The API contract is available [here](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

## Using the Lineage API

The Lineage API allows the clients to query the data flow relationship between different assets managed by Compass.

See the swagger definition of [Lineage API](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json)) for more information.

Lineage API returns a list of directed edges. For each edge, there are `source` and `target` fields that represent nodes to indicate the direction of the edge. Each edge could have an optional property in the `props` field.

Here's a sample API call:

```text
curl http://localhost:8080/v1beta1/lineage/data-project%3Adatalake.events --header 'Compass-User-UUID:odpf@email.com' 

{
    data: [
        {
            "source": {
                "urn": "data-project:datalake.events",
                "type": "table",
                "service": "bigquery",
            },
            "target": {
                "urn": "events-transform-dwh",
                "type": "csv",
                "service": "s3",
            },
            "props": nil
        },
        {
            "source": {
                "urn": "events-ingestion",
                "type": "topic",
                "service": "beast",
            },
            "target": {
                "urn": "data-project:datalake.events",
                "type": "table",
                "service": "bigquery",
            },
            "props": nil
        },
    ]
}
```

The lineage is fetched from the perspective of an asset. The response shows it has a list of upstreams and downstreams assets of the requested asset.
Notice that in the URL, we are using `urn` instead of `id`. The reason is because we use `urn` as a main identifier in our lineage storage. We don't use `id` to store the lineage as a main identifier, because `id` is internally auto generated and in lineage, there might be some assets that we don't store in our Compass' storage yet.



## Using the Upsert Patch API

Upsert Patch API is the one and only ingestion API in Compass. The API contract is available [here](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

If there is an existing asset, Upsert Patch API will check each field whether there is an update in the field of the existing asset. With this behaviour, it is possible to send partial updated field to update a certain field only as long as the `urn`, `type`, and `service` match with the existing asset. If there is any field changed, a new version of the asset will be created.

If the asset does not exist, upsert patch API will create a new asset.

Apart from asset details, we also could send upstreams and downstreams of lineage edges of the asset in the body.

Here is the example of sending updated field with upsert patch API:
```text
$ curl --request PATCH http://localhost:8080/v1beta1/assets --header 'Compass-User-UUID:odpf@email.com'
--data-raw '{
    "asset": {
        "urn": "kafka::booking-sample",
        "type": "topic",
        "name": "sample-name-topic",
        "service": "kafka",
        "description": "sample description",
        "data": {},
        "owners": [
            {
                "id": "",
                "email": "user-owner@email.com"
            }
        ],
        "labels": {}
    },
    "upstreams": [
        {
            "urn": "sample-name-topic",
            "type": "topic",
            "service": "kafka"
        },
        {
            "urn": "sample-urn",
            "type": "table",
            "service": "bigquery"
        }
    ],
    "downstreams": [
        {
            "urn": "sample-urn",
            "type": "table",
            "service": "bigquery"
        },
        {
            "urn": "sample-name-topic",
            "type": "topic",
            "service": "kafka"
        }
    ]
}'
```