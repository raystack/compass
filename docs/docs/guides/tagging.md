# Tagging

This doc explains how to tag an asset in Compass with a specific tag.

## Tag Template
To support reusability of a tag, Compass has a tag template that we need to define first before we apply it to an asset. Tagging an asset means Compass will wire tag template to assets.

Creating a tag's template could be done with Tag Template API.

```bash
$ curl --request POST 'localhost:8080/v1beta1/tags/templates' \
--header 'Compass-User-UUID: user@odpf.io' \
--data-raw '{
    "urn": "my-first-template",
    "display_name": "My First Template",
    "description": "This is my first template",
    "fields": [
        {
            "urn": "fieldA",
            "display_name": "Field A",
            "description": "This is Field A",
            "data_type": "string",
            "required": false
        },
        {
            "urn": "fieldB",
            "display_name": "Field B",
            "description": "This is Field B",
            "data_type": "double",
            "required": true
        }
    ]
}'
```

We can verify the tag's template is created by calling GET tag's templates API

```bash
$ curl --request GET 'localhost:8080/v1beta1/tags/templates' \
--header 'Compass-User-UUID: user@odpf.io'
```
The response will be like this
```javascript
{
    "data": [
        {
            "urn": "my-first-template",
            "display_name": "My First Template",
            "description": "This is my first template",
            "fields": [
                {
                    "id": 1,
                    "urn": "fieldA",
                    "display_name": "Field A",
                    "description": "This is Field A",
                    "data_type": "string",
                    "created_at": "2022-05-10T09:34:18.766125Z",
                    "updated_at": "2022-05-10T09:34:18.766125Z"
                },
                {
                    "id": 2,
                    "urn": "fieldB",
                    "display_name": "Field B",
                    "description": "This is Field B",
                    "data_type": "double",
                    "required": true,
                    "created_at": "2022-05-10T09:34:18.766125Z",
                    "updated_at": "2022-05-10T09:34:18.766125Z"
                }
            ],
            "created_at": "2022-05-10T09:34:18.766125Z",
            "updated_at": "2022-05-10T09:34:18.766125Z"
        }
    ]
}
```

Now, we already have a template with template urn `my-first-template` that has 2 kind of fields with id `1` and `2`.
## Tagging an Asset
Once templates exist, we can tag an asset with a template by calling PUT `/v1beta1/tags/assets/{asset_id}` API.

Assuming we have an asset
```javascript
{
    "id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",
    "urn": "sample-urn",
    "type": "topic",
    "service": "bigquery",
    "name": "sample-name",
    "description": "sample description",
    "version": "0.1",
    "updated_by": {
        "uuid": "user@odpf.io"
    },
    "created_at": "2022-05-11T07:03:45.954387Z",
    "updated_at": "2022-05-11T07:03:45.954387Z"
}
```

We can tag the asset with template `my-first-template`.
```bash
$ curl --request POST 'localhost:8080/v1beta1/tags/assets' \
--header 'Compass-User-UUID: user@odpf.io'
--data-raw '{
    "asset_id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",
    "template_urn": "my-first-template",
    "tag_values": [
        {
            "field_id": 1,
            "field_value": "test"
        },
        {
            "field_id": 2,
            "field_value": 10.0
        }
    ]
}'
```

We will get response showing that the asset is already tagged.
```javascript
{
    "data": {
        "asset_id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",
        "template_urn": "my-first-template",
        "tag_values": [
            {
                "field_id": 1,
                "field_value": "test",
                "field_urn": "fieldA",
                "field_display_name": "Field A",
                "field_description": "This is Field A",
                "field_data_type": "string",
                "created_at": "2022-05-11T00:06:26.475943Z",
                "updated_at": "2022-05-11T00:06:26.475943Z"
            },
            {
                "field_id": 2,
                "field_value": 10,
                "field_urn": "fieldB",
                "field_display_name": "Field B",
                "field_description": "This is Field B",
                "field_data_type": "double",
                "field_required": true,
                "created_at": "2022-05-11T00:06:26.475943Z",
                "updated_at": "2022-05-11T00:06:26.475943Z"
            }
        ],
        "template_display_name": "My First Template",
        "template_description": "This is my first template"
    }
}
``` 

## Getting Asset's Tag(s)
We can get all tags belong to an asset by calling GET `/v1beta1/tags/assets/{asset_id}` API.

```bash
$ curl --request GET 'localhost:8080/v1beta1/tags/assets/a2c74793-b584-4d20-ba2a-28bdf6b92c08' \
--header 'Compass-User-UUID: user@odpf.io'

{
    "data": [
        {
            "asset_id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",
            "template_urn": "my-first-template",
            "tag_values": [
                {
                    "field_id": 1,
                    "field_value": "test",
                    "field_urn": "fieldA",
                    "field_display_name": "Field A",
                    "field_description": "This is Field A",
                    "field_data_type": "string",
                    "created_at": "2022-05-11T00:06:26.475943Z",
                    "updated_at": "2022-05-11T00:06:26.475943Z"
                },
                {
                    "field_id": 2,
                    "field_value": 10,
                    "field_urn": "fieldB",
                    "field_display_name": "Field B",
                    "field_description": "This is Field B",
                    "field_data_type": "double",
                    "field_required": true,
                    "created_at": "2022-05-11T00:06:26.475943Z",
                    "updated_at": "2022-05-11T00:06:26.475943Z"
                }
            ],
            "template_display_name": "My First Template",
            "template_description": "This is my first template"
        }
    ]
}