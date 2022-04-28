# Internals

This document details information about how Compass interfaces with elasticsearch. It is meant to give an overview of how some concepts work internally, to help streamline understanding of how things work under the hood.

## Index Setup

There is a migration command in compass to setup all storages. Once the migration is executed, all types are being created (if does not exist). When a type is created, an index is created in elasticsearch by it's name. All created indices are aliased to the `universe` index, which is used to run the search when all types need to be searched, or when `filter.type` is not specifed in the Search API.

The indices are also configured with a camel case tokenizer, to support proper lexing of some resources that use camel case in their nomenclature \(protobuf names for instance\). Given below is a sample of the index settings that are used:

```javascript
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

```text
curl http://localhost:3000/v1beta1/search?text=log&filter.landscape=id
```

is internally translated to the following elasticsearch query

```javascript
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
