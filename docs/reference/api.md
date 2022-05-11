# API
Documentation of our Compass API with gRPC and gRPC-Gateway.

## Version: 0.2.1

**License:** [Apache License 2.0](https://github.com/odpf/compass/blob/main/LICENSE)

[More about Compass](https://odpf.gitbook.io/compass/)
### /v1beta1/assets

#### GET
##### Summary

Get list of assets

##### Description

Returns list of assets, optionally filtered by types, services, sorting, fields in asset.data and querying fields

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| q | query | filter by specific query | No | string |
| qFields | query | filter by multiple query fields | No | string |
| types | query | filter by multiple types | No | string |
| services | query | filter by multiple services | No | string |
| sort | query | sorting based on fields | No | string |
| direction | query | sorting direction can either be asc or desc  | No | string |
| size | query | maximum size to fetch | No | long |
| offset | query | offset to fetch from | No | long |
| withTotal | query | if set include total field in response | No | boolean |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAllAssetsResponse](#getallassetsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PUT
##### Summary

Upsert an asset

##### Description

Create a new asset if a combination of urn, type and service does not exist. If exists update instead

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [UpsertAssetRequest](#upsertassetrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [UpsertAssetResponse](#upsertassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PATCH
##### Summary

Patch/Create an asset

##### Description

Similar to Upsert but with patch strategy and different body format

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [UpsertPatchAssetRequest](#upsertpatchassetrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [UpsertPatchAssetResponse](#upsertpatchassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/assets/{id}

#### GET
##### Summary

Find an asset

##### Description

Returns a single asset with given ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAssetByIDResponse](#getassetbyidresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### DELETE
##### Summary

Delete an asset

##### Description

Delete a single asset with given ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [DeleteAssetResponse](#deleteassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/assets/{id}/stargazers

#### GET
##### Summary

Find users that stars an asset

##### Description

Returns a list of users that stars an asset

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAssetStargazersResponse](#getassetstargazersresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/assets/{id}/versions

#### GET
##### Summary

Get version history of an asset

##### Description

Returns a list of asset version history

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAssetVersionHistoryResponse](#getassetversionhistoryresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/assets/{id}/versions/{version}

#### GET
##### Summary

Get asset's previous version

##### Description

Returns a specific version of an asset

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| version | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAssetByVersionResponse](#getassetbyversionresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/discussions

#### GET
##### Summary

Get all discussions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| type | query |  | No | string |
| state | query |  | No | string |
| owner | query |  | No | string |
| assignee | query |  | No | string |
| asset | query |  | No | string |
| labels | query |  | No | string |
| sort | query |  | No | string |
| direction | query |  | No | string |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAllDiscussionsResponse](#getalldiscussionsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### POST
##### Summary

Create a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [CreateDiscussionRequest](#creatediscussionrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CreateDiscussionResponse](#creatediscussionresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/discussions/{discussionId}/comments

#### GET
##### Summary

Get all comments of a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| discussionId | path |  | Yes | string |
| sort | query |  | No | string |
| direction | query |  | No | string |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAllCommentsResponse](#getallcommentsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### POST
##### Summary

Create a comment of a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| discussionId | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CreateCommentResponse](#createcommentresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/discussions/{discussionId}/comments/{id}

#### GET
##### Summary

Get a comment of a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| discussionId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetCommentResponse](#getcommentresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### DELETE
##### Summary

Delete a comment of a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| discussionId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [DeleteCommentResponse](#deletecommentresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PUT
##### Summary

Update a comment of a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| discussionId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [UpdateCommentResponse](#updatecommentresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/discussions/{id}

#### GET
##### Summary

Get a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetDiscussionResponse](#getdiscussionresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PATCH
##### Summary

Patch a discussion

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [PatchDiscussionResponse](#patchdiscussionresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/lineage/{urn}

#### GET
##### Summary

Get Lineage Graph

##### Description

Returns the lineage graph. Each entry in the graph describes a (edge) directed relation of assets with source and destination using it's urn, type, and service.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| urn | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetGraphResponse](#getgraphresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/me/discussions

#### GET
##### Summary

Get all discussions of a user

##### Description

Returns all discussions given possible filters of a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| filter | query |  | No | string |
| type | query |  | No | string |
| state | query |  | No | string |
| asset | query |  | No | string |
| labels | query |  | No | string |
| sort | query |  | No | string |
| direction | query |  | No | string |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetMyDiscussionsResponse](#getmydiscussionsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/me/starred

#### GET
##### Summary

Get my starred assets

##### Description

Get all assets starred by me

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetMyStarredAssetsResponse](#getmystarredassetsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/me/starred/{assetId}

#### GET
##### Summary

Get my starred asset

##### Description

Get an asset starred by me

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetMyStarredAssetResponse](#getmystarredassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### DELETE
##### Summary

Unstar an asset

##### Description

Unmark my starred asset

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [UnstarAssetResponse](#unstarassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PUT
##### Summary

Star an asset

##### Description

Mark an asset with a star

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [StarAssetResponse](#starassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/search

#### GET
##### Summary

Search for an asset

##### Description

API for querying documents. 'text' is fuzzy matched against all the available datasets, and matched results are returned. You can specify additional match criteria using 'filter[.*]' query parameters. You can specify each filter multiple times to specify a set of values for those filters. For instance, to specify two landscape 'vn' and 'th', the query could be `/search/?text=<text>&filter[environment]=integration&filter[landscape]=vn&filter[landscape]=th`. As an alternative, this API also supports fuzzy filter match with 'query' query params. For instance, searching assets that has 'bigqu' term in its description `/search/?text=<text>&query[description]=bigqu`

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| text | query | text to search for (fuzzy) | No | string |
| rankby | query | descendingly sort based on a numeric field in the record. the nested field is written with period separated field name. eg, "rankby[data.profile.usage_count]" | No | string |
| size | query | number of results to return | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [SearchAssetsResponse](#searchassetsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/search/suggest

#### GET
##### Summary

Suggest an asset

##### Description

API for retreiving N number of asset names that similar with the `text`. By default, N = 5 for now and hardcoded in the code.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| text | query | text to search for suggestions | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [SuggestAssetsResponse](#suggestassetsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/tags/assets

#### POST
##### Summary

Tag an asset

##### Description

Tag an asset with a tag template

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [CreateTagAssetRequest](#createtagassetrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CreateTagAssetResponse](#createtagassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/tags/assets/{assetId}

#### GET
##### Summary

Get an asset's tags

##### Description

Get all tags for an assets

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAllTagsByAssetResponse](#getalltagsbyassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/tags/assets/{assetId}/templates/{templateUrn}

#### GET
##### Summary

Find a tag by asset and template

##### Description

Find a single tag using asset id and template urn

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |
| templateUrn | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetTagsByAssetAndTemplateResponse](#gettagsbyassetandtemplateresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### DELETE
##### Summary

Remove a tag on an asset

##### Description

Remove a tag on an asset in a type

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |
| templateUrn | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [DeleteTagAssetResponse](#deletetagassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PUT
##### Summary

Update a tag on an asset

##### Description

Update a tag on an asset

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| assetId | path |  | Yes | string |
| templateUrn | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [UpdateTagAssetResponse](#updatetagassetresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/tags/templates

#### GET
##### Summary

Get all tag templates

##### Description

Get all available tag templates

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| urn | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAllTagTemplatesResponse](#getalltagtemplatesresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### POST
##### Summary

Create a template

##### Description

Create a new tag template

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [CreateTagTemplateRequest](#createtagtemplaterequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CreateTagTemplateResponse](#createtagtemplateresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/tags/templates/{templateUrn}

#### GET
##### Summary

Get a tag template

##### Description

Get a single tag template

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| templateUrn | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetTagTemplateResponse](#gettagtemplateresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### DELETE
##### Summary

Delete a tag template

##### Description

Delete a single tag template

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| templateUrn | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [DeleteTagTemplateResponse](#deletetagtemplateresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

#### PUT
##### Summary

Update a template

##### Description

Update an existing tag template

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| templateUrn | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [UpdateTagTemplateResponse](#updatetagtemplateresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/types

#### GET
##### Summary

fetch all types

##### Description

Fetch all types supported in Compass

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetAllTypesResponse](#getalltypesresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### /v1beta1/users/{userId}/starred

#### GET
##### Summary

Get assets starred by a user

##### Description

Get all assets starred by a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path |  | Yes | string |
| size | query |  | No | long |
| offset | query |  | No | long |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GetUserStarredAssetsResponse](#getuserstarredassetsresponse) |
| 400 | Returned when the data that user input is wrong. | [Status](#status) |
| 404 | Returned when the resource does not exist. | [Status](#status) |
| 409 | Returned when the resource already exist. | [Status](#status) |
| 500 | Returned when theres is something wrong on the server side. | [Status](#status) |
| default | An unexpected error response. | [Status](#status) |

### Models

#### Any

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| @type | string |  | No |

#### Asset

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| urn | string |  | No |
| type | string |  | No |
| service | string |  | No |
| name | string |  | No |
| description | string |  | No |
| data | object |  | No |
| labels | object |  | No |
| owners | [ [User](#user) ] |  | No |
| version | string |  | No |
| updatedBy | [User](#user) |  | No |
| changelog | [ [Change](#change) ] |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### Change

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| type | string |  | No |
| path | [ string ] |  | No |
| from | object |  | No |
| to | object |  | No |

#### Comment

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| discussionId | string |  | No |
| body | string |  | No |
| owner | [User](#user) |  | No |
| updatedBy | [User](#user) |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### CreateCommentResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### CreateDiscussionRequest

Request to be sent to create a discussion

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| title | string |  | Yes |
| body | string |  | Yes |
| type | string |  | No |
| state | string |  | No |
| labels | [ string ] |  | No |
| assets | [ string ] |  | No |
| assignees | [ string ] |  | No |

#### CreateDiscussionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### CreateTagAssetRequest

Request to be sent to create a tag

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| assetId | string |  | Yes |
| templateUrn | string |  | Yes |
| tagValues | [ [TagValue](#tagvalue) ] |  | Yes |
| templateDisplayName | string |  | No |
| templateDescription | string |  | No |

#### CreateTagAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [v1beta1.Tag](#v1beta1tag) |  | No |

#### CreateTagTemplateRequest

Request to be sent to create a tag's template

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| urn | string |  | Yes |
| displayName | string |  | Yes |
| description | string |  | Yes |
| fields | [ [TagTemplateField](#tagtemplatefield) ] |  | No |

#### CreateTagTemplateResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [TagTemplate](#tagtemplate) |  | No |

#### DeleteAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| DeleteAssetResponse | object |  |  |

#### DeleteCommentResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| DeleteCommentResponse | object |  |  |

#### DeleteTagAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| DeleteTagAssetResponse | object |  |  |

#### DeleteTagTemplateResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| DeleteTagTemplateResponse | object |  |  |

#### Discussion

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| title | string |  | No |
| body | string |  | No |
| type | string |  | No |
| state | string |  | No |
| labels | [ string ] |  | No |
| assets | [ string ] |  | No |
| assignees | [ string ] |  | No |
| owner | [User](#user) |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### GetAllAssetsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Asset](#asset) ] |  | No |
| total | long |  | No |

#### GetAllCommentsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Comment](#comment) ] |  | No |

#### GetAllDiscussionsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Discussion](#discussion) ] |  | No |

#### GetAllTagTemplatesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [TagTemplate](#tagtemplate) ] |  | No |

#### GetAllTagsByAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [v1beta1.Tag](#v1beta1tag) ] |  | No |

#### GetAllTypesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [v1beta1.Type](#v1beta1type) ] |  | No |

#### GetAssetByIDResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [Asset](#asset) |  | No |

#### GetAssetByVersionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [Asset](#asset) |  | No |

#### GetAssetStargazersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [User](#user) ] |  | No |

#### GetAssetVersionHistoryResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Asset](#asset) ] |  | No |

#### GetCommentResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [Comment](#comment) |  | No |

#### GetDiscussionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [Discussion](#discussion) |  | No |

#### GetGraphResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [LineageEdge](#lineageedge) ] |  | No |

#### GetMyDiscussionsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Discussion](#discussion) ] |  | No |

#### GetMyStarredAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [Asset](#asset) |  | No |

#### GetMyStarredAssetsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Asset](#asset) ] |  | No |

#### GetTagTemplateResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [TagTemplate](#tagtemplate) |  | No |

#### GetTagsByAssetAndTemplateResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [v1beta1.Tag](#v1beta1tag) |  | No |

#### GetUserStarredAssetsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Asset](#asset) ] |  | No |

#### LineageEdge

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| source | string |  | No |
| target | string |  | No |
| prop | object |  | No |

#### LineageNode

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| urn | string |  | No |
| type | string |  | No |
| service | string |  | No |

#### NullValue

`NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

- NULL_VALUE: Null value.

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| NullValue | string | `NullValue` is a singleton enumeration to represent the null value for the `Value` type union.   The JSON representation for `NullValue` is JSON `null`.   - NULL_VALUE: Null value. |  |

#### PatchDiscussionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| PatchDiscussionResponse | object |  |  |

#### SearchAssetsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ [Asset](#asset) ] |  | No |

#### StarAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### Status

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| message | string |  | No |
| details | [ [Any](#any) ] |  | No |

#### SuggestAssetsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [ string ] |  | No |

#### TagTemplate

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| urn | string |  | No |
| displayName | string |  | No |
| description | string |  | No |
| fields | [ [TagTemplateField](#tagtemplatefield) ] |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### TagTemplateField

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | long |  | No |
| urn | string |  | No |
| displayName | string |  | No |
| description | string |  | No |
| dataType | string |  | No |
| options | [ string ] |  | No |
| required | boolean |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### TagValue

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| fieldId | long |  | No |
| fieldValue | object |  | No |
| fieldUrn | string |  | No |
| fieldDisplayName | string |  | No |
| fieldDescription | string |  | No |
| fieldDataType | string |  | No |
| fieldOptions | [ string ] |  | No |
| fieldRequired | boolean |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### UnstarAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| UnstarAssetResponse | object |  |  |

#### UpdateCommentResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| UpdateCommentResponse | object |  |  |

#### UpdateTagAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [v1beta1.Tag](#v1beta1tag) |  | No |

#### UpdateTagTemplateResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data | [TagTemplate](#tagtemplate) |  | No |

#### UpsertAssetRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset | [UpsertAssetRequest.BaseAsset](#upsertassetrequestbaseasset) |  | No |
| upstreams | [ [LineageNode](#lineagenode) ] |  | No |
| downstreams | [ [LineageNode](#lineagenode) ] |  | No |

#### UpsertAssetRequest.BaseAsset

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| urn | string |  | No |
| type | string |  | No |
| name | string |  | No |
| service | string |  | No |
| description | string | description of an asset | No |
| data | object |  | No |
| labels | object | labels of an asset | No |
| owners | [ [User](#user) ] | list of owners of the asset | No |

#### UpsertAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### UpsertPatchAssetRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset | [UpsertPatchAssetRequest.BaseAsset](#upsertpatchassetrequestbaseasset) |  | No |
| upstreams | [ [LineageNode](#lineagenode) ] |  | No |
| downstreams | [ [LineageNode](#lineagenode) ] |  | No |

#### UpsertPatchAssetRequest.BaseAsset

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| urn | string |  | No |
| type | string |  | No |
| name | string | name of an asset | No |
| service | string |  | No |
| description | string | description of an asset | No |
| data | object | dynamic data of an asset | No |
| labels | object | labels of an asset | No |
| owners | [ [User](#user) ] | list of owners of the asset | No |

#### UpsertPatchAssetResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### User

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| uuid | string |  | No |
| email | string |  | No |
| provider | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1.Tag

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| assetId | string |  | No |
| templateUrn | string |  | No |
| tagValues | [ [TagValue](#tagvalue) ] |  | No |
| templateDisplayName | string |  | No |
| templateDescription | string |  | No |

#### v1beta1.Type

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| count | long |  | No |
