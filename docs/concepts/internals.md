# Internals

This document details information about how Columbus interfaces with elasticsearch. It is meant to give an overview of how some concepts work internally, to help streamline understanding of how things work under the hood.

## Index Setup

When an type is created, an index is created in elasticsearch by it's name. All created indices are aliased to the `universe` index, which is used to run the search when all types need to be searched, or when `filter.type` is not specifed in the Search API.

The indices are also configured with a camel case tokenizer, to support proper lexing of some resources that use camel case in their nomenclature (protobuf names for instance). Given below is a sample of the index settings that are used:
```json
// PUT http://${ES_HOST}/{index}
{
		"mappings": {},         // used for boost
		"aliases": {            // all indices are aliased to the "universe" index
			"universe": {} 
		},
		"settings": {           // configuration for handling camel case text
			"analysis": {
				"analyzer": {
					"default": {
						"type": "pattern",
						"pattern": "([^\\p{L}\\d]+)|(?<=\\D)(?=\\d)|(?<=\\d)(?=\\D)|(?<=[\\p{L}&&[^\\p{Lu}]])(?=\\p{Lu})|(?<=\\p{Lu})(?=\\p{Lu}[\\p{L}&&[^\\p{Lu}]])"
					}
				}
			}
		}
	}
```

## Search

We use elasticsearch's `multi_match` search for running our queries. Depending on whether there are additional filter's specified during search, we augument the query with a custom script query that filter's the result set. 

The script filter is designed to match a document if:
* the document contains the filter key and it's value matches the filter value OR
* the document doesn't contain the filter key at all


To demonstrate, the following API call:
```
curl http://localhost:3000/v1/search?text=log&filter.landscape=id
```
is internally translated to the following elasticsearch query
```json
// GET http://${ES_HOST}/universe/_search
{
    "query": {
        "bool": {
            "must": {
                "multi_match": {
                    "query": "log"
                }
            },
            "filter": [{
                "script": {
                    "script": {
                        "source": "doc.containsKey(\"landscape.keyword\") == false || doc[\"landscape.keyword\"].value == \"id\""
                    }
                }
            }]
        }
    }
}
```

## Lineage

The process of building Lineage's graph is considered heavy and consumes more memories as more resource and types are added.

The building process can be expressed using the following pseudo-code:
```js
graph = {} // an empty map
all_types = fetch_all_types()
for typ in all_types
    all_type_resources = fetch_all_resources(typ)
    for resource in all_type_resources
        entry = populate_upstreams(resource)
        add_entry_to_graph(entry)
   for entry in graph:
        populate_downstreams(entry)
```

From pseudo-code above we can see that we practically fetch all records and types from Elastisearch then process all of them to build the lineage graph.

Doing it for every request would consume a lot of computing power and memory, especially if you have a lot of records stored. So instead of building it for each lineage request, Columbus will builds lineage graph on two conditions:
* on app start
* a configurable time interval (default to 5 minutes) after previous building process

The lineage graph is then cached in memory by adding or replacing the previous graph.

This cached graph will then be used to serve lineage request until it is being replaced by a newly built graph.
