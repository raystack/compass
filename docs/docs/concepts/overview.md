# Overview

Compass has three major concept when it comes to data ingestion: Asset, Type, and Service.

Asset is essentially an arbitrary JSON object that represent a metadata of a specific service with a specific type.

Type defines a ‘type’ of an asset and it is pre-defined. There are currently 4 supported types in Compass: `table`, `job`, `dashboard`, and `topic`.
 
Service defines the application name that the asset was coming from. For example: `biquery`, `postgres`, etc. If you wanted to push data for `bigquery` dataset\(s\) to Compass, you would need to first define the ‘`bigquery`’ service in compass.

Some features that compass has:
* Asset Tagging
* User
* Discussion
* Starring

## Asset

An Asset is a JSON document that describes a metadata. Asset has a schema:
|  Field | Required | Type   | Description |
|---|---|---|---|
|  id | false |  string |  compass' auto-generated uuid |
|  urn | true | string |  external identifier of the metadata |
|  type | true | string  |  type of metadata, only supports `table`, `job`, `topic`,`dashboard` |
|  service | true | string  |  application name where the metadata was coming from e.g. `bigquery`, `postgres` |
|  name | true | string  |  name of the metadata |
|  description | false | string  | description of the metadata  |
|  data | false | json |  dynamic data |
|  labels | false |json  |  labels of metadata, written in key-value string  |
|  owners | false | []json | array of json, where each json contains `email` field  |

```text
{

    "urn": "topic/order-log",
    "type": "topic",
    "service": "kafka",
    "description": "desc",
    "data": {
        "some_data1": {
            "random_data": 123,
            "nested_data": {
                "boolean_data": true
            }
        }, 
        "some_data1": "value"
    }
    "labels": {
        "labelkey1": "labelvalue1", 
        "labelkey2": "labelvalue2"
    },
    "users": [
        {
            "email": "user@odpf.io"
        }
    ]
}
```

Every asset that is pushed SHOULD have the required fields: `urn`, `type`, `service`, `name`. The value of these fields MUST be string, if present. 

Asset ingestion API \(/v1beta1/assets\) is using HTTP PATCH method. The behavioud would be similar with how PATCH works. It is possible to patch one field only in an asset by sending the updated field to the ingestion API. This also works for the data in dynamic `data` field. The combination of `urn`, `type`, `service` will be the identifier to patch an asset.
In case the `urn` does not exist, the asset ingestion PATCH API \(/v1beta1/assets\) will create a new asset.

## Lineage

Lineage is the origin or history of an asset. It represents a series of transformation of one or many assets.

Each asset can have downstream/s and upstream/s. An asset without a single downstream, tells us that it is the end of the lineage, while an asset without a single upstream means that it is a start of a lineage.

This is how a lineage is currently being represented
```text
[
    {
        "source": {
            "urn": "topic/order-log",
            "type": "topic",
            "service": "kafka"
        },
        "target": {
            "urn": "bqtable/order_monthly",
            "type": "table",
            "service": "bigquery"
        },
        "props": nil
    },    
    {
        "source": {
            "urn": "topic/order-log",
            "type": "topic",
            "service": "kafka"
        },
        "target": {
            "urn": "bqtable/order_daily",
            "type": "table",
            "service": "bigquery"
        },
        "props": nil
    },
]
```

## Asset Versioning
Compass versions each updated asset ingested via Upsert Patch API. The base version of an asset is `v0.1`. The base version will be given to the newly created asset. If there is any changes in the asset schema, a new version will be created. 
Up until now, Compass always bump up the minor version if an asset get updated. The version history of an asset could also be fetched via [/v1beta1/assets/{id}/versions](https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json) API.
Not only storing the versions of an asset, Compass also stores the changelog between each version. Compass use [r3labs/diff](https://github.com/r3labs/diff) to get the diff between newly ingested asset and the existing asset.

For instance, there is an asset with urn `kafka:booking-log-kafka`
```text
{
    "id": "f2bb4e02-12b6-4c9f-aa9d-7d56aaaeb51e",
    "urn": "kafka:booking-log-kafka",
    "type": "topic",
    "service": "kafka",
    "data": {},
    "labels": {
        "environment": "integration"
    },
    "version": "0.1"
}
```

If there is an update to the `environment` in the asset labels, here is the asset version history stored in Compass:
```text
{
    "id": "f2bb4e02-12b6-4c9f-aa9d-7d56aaaeb51e",
    "urn": "kafka:booking-log-kafka",
    "type": "topic",
    "service": "kafka",
    "data": {},
    "labels": {
        "environment": "production"
    },
    "version": "0.2"
    "changelog": [
        {
            "type": "update",
            "path": ["labels","environment"],
            "from": "integration",
            "to":   "production
        }
    ]
}
```

## Tagging an Asset
Compass allows user to tag a specific asset. To tag a new asset, one needs to create a template of the tag. Tag's template defines a set of fields' tag that are applicable to tag each field in an asset.
Once a template is created, each field in an asset is possible to be tagged by calling `/v1beta1/tags` API. More detail about [Tagging](../guides/tagging.md).

## User
The current version of Compass does not have user management. Compass expect there is an external instance that manages user. Compass consumes user information from the configurable identity uuid header in every API call. The default name of the header is `Compass-User-UUID`. 
Compass does not make any assumption of what kind of identity format that is being used. The `uuid` indicates that it could be in any form (e.g. email, UUIDv4, etc) as long as it is universally unique.
The current behaviour is, Compass will add a new user if the user information consumed from the header does not exist in Compass' database. More detail about [User](./user.md).
## Discussion
Compass supports discussion feature. User could drop comments in each discussion. Currently, there are three types of discussions `issues`, `open ended`, and `question and answer`. Depending on the type, the discussion could have multiple possible states. In the current version, all types only have two states: `open` and `closed`. A newly created discussion will always be assign an `open` state. More detail about [Discussion](../guides/discussion.md).

## Starring
Compass allows a user to stars an asset. This bookmarking functionality is introduced to increase the speed of a user to get information. There is also an API to see which users star an asset (stargazers). More detail about [Starring](../guides/starring.md).