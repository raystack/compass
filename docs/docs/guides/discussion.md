import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Discussion

Discussion is a new feature in Compass. One could create a discussion and all users can put comment in it. Currently, there are three types of discussions `issues`, `open ended`, and `question and answer`. Depending on the type, the discussion could have multiple possible states. In the current version, all types only have two states: `open` and `closed`. A newly created discussion will always be assign an `open` state.

## Create a Discussion

A discussion thread can be created with the Discussion API. The API contract is available [here](https://github.com/goto/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass discussion post --body=<filepath to discussion body>
```

```json
{
  "title": "The first discussion",
  "body": "This is the first discussion thread in Compass",
  "type": "openended"
}
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl --request POST 'http://localhost:8080/v1beta1/discussions' \
--header 'Compass-User-UUID:gotocompany@email.com' \
--data-raw '{
  "title": "The first discussion",
  "body": "This is the first discussion thread in Compass",
  "type": "openended"
}'
```
</TabItem>
</Tabs>

## Fetching All Discussions

The Get Discussions will fetch all discussions in Compass.

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass discussion list
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/discussions' \
--header 'Compass-User-UUID:gotocompany@email.com'
```
</TabItem>
</Tabs>

The response will be something like
```javascript
{
    "data": [
        {
            "id": "1",
            "title": "The first discussion",
            "body": "This is the first discussion thread in Compass",
            "type": "openended"
            "state": "open",
            "labels": [],
            "assets": [],
            "assignees": [],
            "owner": {
                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",
                "email": "gotocompany@email.com",
                "provider": "shield"
            },
            "created_at": "elit cillum Duis",
            "updated_at": "velit dolor ex"
        }
    ]
}
```
Notice the state is `open` by default once we create a new discussion. There are also some additional features in discussion where we can label the discussion and assign users and assets to the discussion. These labelling and assinging assets and users could also be done when we are creating a discussion.

## Patching Discussion

If we are not labelling and assigning users & assets to the discussion in the creation step, there are also a dedicated API to do those.

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/discussions/1' \
--header 'Compass-User-UUID:gotocompany@email.com' \
--data-raw '{
    "title": "The first discussion (duplicated)",
    "state": "closed"
}'
```

We just need to send the fields that we want to patch for a discussion. Some fields have array type, in this case the PATCH will overwrite the fields with the new value.

For example we have this labelled discussion.
```bash
$ curl 'http://localhost:8080/v1beta1/discussions' \
--header 'Compass-User-UUID:gotocompany@email.com'

{
    "data": [
        {
            "id": "1",
            "title": "The first discussion",
            "body": "This is the first discussion thread in Compass",
            "type": "openended"
            "state": "open",
            "labels": [
                "work",
                "urgent",
                "help wanted"
            ],
            "owner": {
                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",
                "email": "gotocompany@email.com",
                "provider": "shield"
            },
            "created_at": "elit cillum Duis",
            "updated_at": "velit dolor ex"
        }
    ]
}
```

If we patch the label with the new values.

```bash
$ curl --request PATCH 'http://localhost:8080/v1beta1/discussions/1' \
--header 'Compass-User-UUID:gotocompany@email.com' \
--data-raw '{
    "labels": ["new value"]
}'
```

The discussion with id 1 will be updated like this.
```bash
$ curl 'http://localhost:8080/v1beta1/discussions' \
--header 'Compass-User-UUID:gotocompany@email.com'

{
    "data": [
        {
            "id": "1",
            "title": "The first discussion",
            "body": "This is the first discussion thread in Compass",
            "type": "openended"
            "state": "open",
            "labels": [
                "new value"
            ],
            "owner": {
                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",
                "email": "gotocompany@email.com",
                "provider": "shield"
            },
            "created_at": "elit cillum Duis",
            "updated_at": "velit dolor ex"
        }
    ]
}
```

## Commenting a Discussion

One could also comment a specific discussion with discussion comment API.

```bash
$ curl --request POST 'http://localhost:8080/v1beta1/discussions/1/comments' \
--header 'Compass-User-UUID:gotocompany@email.com' \
--data-raw '{
  "body": "This is the first comment of discussion 1"
}'
```

## Getting All My Discussions

Compass integrates discussions with User API so we could fetch all discussions belong to us with this API.
```bash
$ curl 'http://localhost:8080/v1beta1/me/discussions' \
--header 'Compass-User-UUID:gotocompany@email.com'

{
    "data": [
        {
            "id": "1",
            "title": "The first discussion",
            "body": "This is the first discussion thread in Compass",
            "type": "openended"
            "state": "open",
            "labels": [
                "new value"
            ],
            "owner": {
                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",
                "email": "gotocompany@email.com",
                "provider": "shield"
            },
            "created_at": "elit cillum Duis",
            "updated_at": "velit dolor ex"
        }
    ]
}
```