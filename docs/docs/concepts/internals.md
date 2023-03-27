# Internals

This document details information about how Compass interfaces with elasticsearch. It is meant to give an overview of how some concepts work internally, to help streamline understanding of how things work under the hood.

## Index Setup

There is a migration command in compass to setup all storages. The indices are configured with a camel case tokenizer, to support proper lexing of some resources that use camel case in their nomenclature \(protobuf names for instance\). Given below is a sample of the index settings that are used:

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

One shared index is created for all services and tenants but each request(read/write) is routed to a unique shard for each tenant. Compass categorize tenants into two tires, `shared` and `dedicated`. For shared tenants, all the requests will be routed by namespace id over a single shard in an index. For dedicated tenants, each tenant will have its own index. Note, a single index will have N number of `types` same as the number of `Services` supported in Compass. This design will ensure, all the document insert/query requests are only confined to a single shard(in case of shared) or a single index(in case of dedicated).
Details on why we did this is available at [issue #208](https://github.com/odpf/compass/issues/208).

## Postgres

To enforce multi-tenant restrictions at the database level, [Row Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) is used. RLS requires Postgres users used for application database connection not to be a table owner or a superuser else all RLS are bypassed by default. That means a Postgres user that is migrating the application and a user that is used to serve the app should both be different.

To create a postgres user
```sql
CREATE USER "compass_user" WITH PASSWORD 'compass';
GRANT CONNECT ON DATABASE "compass" TO "compass_user";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "compass_user";
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO "compass_user";
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO "compass_user";

ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT SELECT, INSERT, UPDATE, DELETE, REFERENCES
ON TABLES TO "compass_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT USAGE ON SEQUENCES TO "compass_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT EXECUTE ON FUNCTIONS TO "compass_user";
```

A middleware for grpc looks for `x-namespace-id` header to extract tenant id if not found falls back to `default` namespace. 
Same could be passed in a `jwt token` of Authentication Bearer with `namespace_id` as a claim.

## Search

We use elasticsearch's `multi_match` search for running our queries. Depending on whether there are additional filter's specified during search, we augment the query with a custom script query that filter's the result set.

The script filter is designed to match a document if:

* the document contains the filter key and it's value matches the filter value OR
* the document doesn't contain the filter key at all

To demonstrate, the following API call:

```text
$ curl http://localhost:8080/v1beta1/search?text=log&filter[landscape]=id
```

is internally translated to the following elasticsearch query

```javascript
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

Compass also supports filter with fuzzy match with `query` query params. The script query is designed to match a document if:

* the document contains the filter key and it's value is fuzzily matches the `query` value

```text
$ curl http://localhost:8080/v1beta1/search?text=log&filter[landscape]=id
```

is internally translated to the following elasticsearch query

```javascript
{
   "query":{
      "bool":{
         "filter":{
            "match":{
               "description":{
                  "fuzziness":"AUTO",
                  "query":"test"
               }
            }
         },
         "should":{
            "bool":{
               "should":[
                  {
                     "multi_match":{
                        "fields":[
                           "urn^10",
                           "name^5"
                        ],
                        "query":"log"
                     }
                  },
                  {
                     "multi_match":{
                        "fields":[
                           "urn^10",
                           "name^5"
                        ],
                        "fuzziness":"AUTO",
                        "query":"log"
                     }
                  },
                  {
                     "multi_match":{
                        "fields":[
                           
                        ],
                        "fuzziness":"AUTO",
                        "query":"log"
                     }
                  }
               ]
            }
         }
      }
   },
   "min_score":0.01
}
```