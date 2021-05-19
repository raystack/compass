# API
Universal Search API

## Version: 0.1.0

### /lineage

#### GET
##### Summary:

lineage list api

##### Description:

Returns the lineage graph, optionally filtered by type. Each entry in the graph describes a resource using it's urn and type, and has `downstreams` and `upstreams` fields that declare related resources. By default, the returned graph will only show immediate and directly related resources. For instance, say that according to the lineage configuration, there exist 3 resources R1,  R2 and R3 where data flows from R1 -> R2 -> R3. If the graph is requested with the filter for R1 and R3 , the returned Graph will have a Node R1 that references a downstream R2, but since it was filtered out, it won't be available in the graph. Similarly, R3 will declare a phamtom upstream R2. This can be addressed via the `collapse` feature. If we make the same request with collapse set to true, R1 will declare R3 as its downstream (using trasitive property) and R3 will also have a corresponding upstream declaration of R1.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| filter.type | query |  | No | string |
| collapse | query |  | No | boolean |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [AdjacencyMap](#adjacencymap) |
| 404 | resource not found | [Error](#error) |

### /lineage/{type}/{resource}

#### GET
##### Summary:

lineage get api

##### Description:

Returns lineage graph of a single resource. For BQTable to BQTable lineage, set collapse to true

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| collapse | query |  | No | boolean |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [AdjacencyMap](#adjacencymap) |
| 404 | invalid type requested | [Error](#error) |

### /types

#### GET
##### Summary:

fetch all types

##### Description:

used to fetch all types

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [Type](#type) ] |

#### PUT
##### Summary:

initialise a type

##### Description:

used for initialising/update a type. A type in columbus's nomenclature is a "collection" of documents that belong to a single named type. Type holds metadata about this collection, used when serving search requests

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
|  | body |  | No | [Type](#type) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | OK | [Type](#type) |
| 400 | invalid type | [Error](#error) |

### /types/{name}

#### PUT
##### Summary:

upload documents for a given type.

##### Description:

Use this API for adding records for a certain type. The document can have any number of fields, however; it must atleast have fields specified by 'title' and 'id' properties on type.record_attributes. The value of these properties must be string and they must be located at the object root.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| name | path |  | Yes | string |
| payload | body |  | No | [ [Record](#record) ] |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [Status](#status) |
| 400 | validation error | [ValidationError](#validationerror) |

#### DELETE
##### Summary:

delete a type by its name.

##### Description:

Use this API to delete a type along with all of its records. This is an idempotent operation.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| name | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | string |
| 422 | reserved type name error | [Error](#error) |

### /types/{name}/details

#### GET
##### Summary:

fetch a type details

##### Description:

used to fetch type details by its name

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [Type](#type) |
| 404 | type not found | [Error](#error) |

### /types/{name}/records

#### GET
##### Summary:

list documents for the type

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| name | path |  | Yes | string |
| filter.environment | query | environment name for filtering the records only for specific environment | No | string |
| select | query | comma separated list of fields to return per record (only toplevel keys are supported) | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [Record](#record) ] |
| 400 | bad input | [Error](#error) |
| 404 | not found | [Error](#error) |

### /types/{name}/records/{id}

#### DELETE
##### Summary:

delete a record in a type by its record ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| name | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | string |
| 404 | type or record cannot be found | [Error](#error) |

### /types/{name}/{id}

#### GET
##### Summary:

get a record by id

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| name | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [Record](#record) |
| 404 | document or type does not exist | [Error](#error) |

### /search

#### GET
##### Summary:

search for resources

##### Description:

API for querying documents. 'text' is fuzzy matched against all the available datasets, and matched results are returned. You can specify additional match criteria using 'filter.*' query parameters. You can specify each filter multiple times to specify a set of values for those filters. For instance, to specify two landscape 'vn' and 'th', the query could be `/search/?text=<text>&filter.environment=integration&filter.landscape=vn&filter.landscape=th`

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| text | query | text to search for (fuzzy) | Yes | string |
| size | query | number of results to return | No | integer |
| filter.environment | query | restrict results to specified environment(s) eg, integrated, test, staging, production | No | string |
| filter.landscape | query | restrict results to specified landscape(s) | No | string |
| filter.entity | query | restrict results to specified organisation | No | string |
| filter.type | query | restrict results to the specified types (as in a Columbus type, for instance "dagger", or "firehose") | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [SearchResult](#searchresult) ] |
| 400 | misconfigured request parameters | [Error](#error) |

### Models


#### Classifications

defines the 'class' of the resource

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| Classifications | string | defines the 'class' of the resource |  |

#### Record

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| urn | string |  | No |
| team | string |  | No |
| environment | string |  | No |

#### Status

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string |  | No |

#### AdjacencyEntry

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| urn | string |  | No |
| type | string |  | No |
| downstreams | [ string ] |  | No |
| upstreams | [ string ] |  | No |

#### AdjacencyMap

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| <NodeLabel> | [AdjacencyEntry](#adjacencyentry) |  | No |

#### Type

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string | name of the type (for e.g. dagger, firehose) | No |
| classification | string | defines the 'class' of the resource | No |
| record_attributes | object | defines metadata for the documents that belong to this type. All properties under record_attributes define(s) a mapping of logical purpose, to the name of the key(s) in the documents that hold those information | No |

#### Error

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| reason | string | error message | No |

#### SearchResult

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | URN of the resource | No |
| title | string | describes the resource in a human readable form | No |
| type | string | the individual type of the resource. For example: dagger, firehose | No |
| description | string | optional description of the record | No |
| classification | string | defines the 'class' of the resource | No |
| labels | object | key value pairs describing the labels configured for the given type of resource. Example of labels: team, created, owner etc | No |

#### ValidationError

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| ValidationError |  |  |  |