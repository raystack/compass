# Concepts

Columbus has two major concepts when it comes to data ingestion: Types and Records. Types define a ‘type’ of resource. For example, if you wanted to push data for bigquery dataset(s) to Columbus, you would need to first define the ‘bigquery’ type in columbus. 

Records are arbitrary JSON objects that define a searchable data-point for an type. These don’t have a fixed schema, but must have certain fields in them. (more on this later)

## Types
‘Types’ define a resource type and its corresponding metadata.  Before you can push data to Columbus for a given resource, you need to first define that resource. You can define an type by making a HTTP PUT call to the /v1/types/ API, with the body containing the type definition.
Here’s an example:
```
curl -XPUT http://${COLUMBUS_HOST} \
    -H 'content-type: application/json' \
    -d '
{
    "name": "firehose",
    "classification": "resource",
    "record_attributes": {
        "id": "urn",
        "title": "title",
        "description": "description",
        "labels": ["owner", "status"]
    },
    "lineage": [
        {
            "type": "topic",
            "direction": "upstream",
            "query": "$.topic"
        }
    ],
    "boost": {
        "title": 1.5
    }
}'
```
Let’s briefly discuss the type definition:
* name defines the type of the resource.
* classification defines the logical ‘type’ for a resource. Currently, we have 3 kinds of classifications: ‘resource’ for application deployments, ‘dataset’ for datastores (like bigquery) and ‘schema’ for protobuf definitions. This is used as a filter criteria when processing search queries.
* record_attributes define the partial schema for a record belonging to the type. 
    * id is the field in the record that is used as the primary ID for a record. (mandatory)
    * title is the field containing the human readable name for this resource (mandatory)
    * description is the field that contains the human readable description
    * labels define a list of arbitrary fields that are returned as a part of the search result. 
* lineage is used to define IO relationship between a certain type and another type. Any type can declare it's lineage with respect to another. In this case, `firehose` type declares that it reads from a `topic`, the ID if which can be found using the `$.topic` json path query on respective firehose documents.
* boost is used for defining the relative relevancy of each field in the type documents. By default, all fields have a boost value of `1.0`. You can change this value to fine-tune the order in which search results show up for a given type.

When Columbus answers a search request, it uses the fields defined in record_attributes to generate the response. This lets you change how the response for a particular type will look, without changing the data (records) for that type. The id field is also used as the primary identifier for a record, and is for referenced making create/replace decisions internally.


## Records
A Record is a JSON document that describes a data-point for an type. The schema for a record is loosely defined as the set of fields referenced in  type.record_attributes. To demonstrate what that means, let’s take the example of a hypothetical type:
```
{
    "name": "imaginary-type",
    "classification": "resource",
    "record_attributes": {
        "id": "urn",
        "title": "name",
        "description": "desc",
        "labels": ["label1", "label2"]
    }
}
```


Every record that is pushed for this Type SHOULD have the fields: urn, name, desc, label1 and label2. The value of these fields MUST be string, if present. Note that fields defined by type.record_attributes.id and type.record_attributes.title are MANDATORY. That means that the following record is valid:


```
{
    "urn": "1",
    "title": "first"
}
```

but this one is not:
```
{
    "title": "first",  // missing "urn"
    "desc": "lorem ipsum",
    "label1": "value1",
    "label2": "value2"
}
```

Record Ingestion API (/v1/types/{name}) is idempotent, and it’s safe to send the same record multiple times. If a record has changed, pushing it to ingestion API will replace the older record.
Every record is produced to some type (or more accurately, every record belongs to some type) but the search is run against every record, irrespective of the type (unless restricted by a filter criteria). 

## Lineage

Lineage is the origin or history of a resource. It represents a series of transformation of one or many data points to create a resource.

Each data point can have downstream/s and upstream/s. A data point without a single downstream, tells us that it is the end of the lineage, while a data point without a single upstream means that it is a start of a lineage.

This is how a lineage is currently being represented
```
{
    "topic/order-log": {
        "type": "topic",
        "urn": "order-log",
        "upstreams": null,
        "downstreams": [
          "bqtable/order_monthly",
          "bqtable/order_daily"
        ]
    },
    "bqtable/order_monthly": {
        "type": "bqtable",
        "urn": "order_monthly",
        "upstreams": [
          "topic/order-log"
        ],
        "downstreams": null
     },
    "bqtable/order_daily": {
        "type": "bqtable",
        "urn": "order_daily",
        "upstreams": [
          "topic/order-log"
        ],
        "downstreams": null
     }
}
```
