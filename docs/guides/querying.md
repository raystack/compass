# Querying  metadata

## Starting the server

Compass interfaces with an elasticsearch cluster. Run compass using:

```text
./compass -elasticsearch-brokers "http://<broker-host-name>"
```

Elasticsearch brokers can alternatively be specified via the `ELASTICSEARCH_BROKERS` environment variable.

If you used Docker to build compass, then configuring networking requires extra steps. Following is one of doing it, running elasticsearch inside docker

```text
# create a docker network where compass and elasticsearch will reside 
$ docker network create compass-net

# run elasticsearch, bound to the network we created. Since we are using the -d flag to docker run, the command inside the subshell returns the container id
$ ES_CONTAINER_ID=$(docker run -d -e "discovery.type=single-node" --net compass-net elasticsearch:7.5.2)

# run compass, passing in the hostname (container id) of the elasticsearch server
# if everything goes ok, you should say something like this:
# time="2020-04-01T18:41:00Z" level=info msg="compass v0.1.0-103-g83b909b starting on 0.0.0.0:8080" reporter=main
# time="2020-04-01T18:41:00Z" level=info msg="connected to elasticsearch cluster \"docker-cluster\" (server version 7.5.2)" reporter=main
$ docker run --net compass-net compass -p 8080:8080 -elasticsearch-brokers http://${ES_CONTAINER_ID}:9200
```

## Using the Search API

The API contract is available here: [http://localhost:3000/swagger.yaml](http://localhost:3000/swagger.yaml)

To demonstrate how to use compass, we’re going to query it for resources that contain the word ‘booking’.

```text
$ curl http://localhost:3000/v1/search?text=booking
```

This will return a list of search results. Here’s a sample response:

```text
[
  {
    "title": "g-godata-id-seg-enriched-booking-dagger",
    "id": "g-godata-id-seg-enriched-booking-dagger",
    "type": "dagger",
    "description": "",
    "labels": {
      "flink_name": "g-godata-id-playground",
      "sink_type": "kafka"
    }
  },
  {
    "title": "g-godata-id-booking-bach-test-dagger",
    "id": "g-godata-id-booking-bach-test-dagger",
    "type": "dagger",
    "description": "",
    "labels": {
      "flink_name": "g-godata-id-playground",
      "sink_type": "kafka"
    }
  },
  {
    "title": "g-godata-id-booking-bach-test-3-dagger",
    "id": "g-godata-id-booking-bach-test-3-dagger",
    "type": "dagger",
    "description": "",
    "labels": {
      "flink_name": "g-godata-id-playground",
      "sink_type": "kafka"
    }
  }
]
```

ID is the URN of the resource, while Title is the human friendly name for it. See the complete API spec to learn more about what the rest of the fields mean.

Compass also supports restricting search results via filters. For instance, to restrict search results to the ‘id’ landscape for ‘odpf’ organisation, run:

$ curl [http://localhost:3000/v1/search?text=booking&filter.landscape=vn&filter.entity=odpf](http://localhost:3000/v1/search?text=booking&filter.landscape=vn&filter.entity=odpf)

Under the hood, filter's work by checking whether the matching document's contain the filter key and checking if their values match. Filters can be specified multiple times to specify a set of filter criteria. For example, to search for ‘booking’ in both ‘vn’ and ‘th’ landscape, run:

```text
$ curl http://localhost:3000/v1/search?text=booking&filter.landscape=id&filter.landscape=th
```

You can also specify the number of maximum results you want compass to return using the ‘size’ parameter

```text
$ curl http://localhost:3000/v1/search?text=booking&size=5
```

## Using the Lineage API

The Lineage API allows the clients to query the data flow relationship between different types \(formerly called entities\) managed by Compass.

See the swagger definition of [Lineage API](http://localhost:3000/swagger.yaml) for more information.

Lineage API returns a hashmap representation of a graph \(called AdjacencyMap\). The values of the hashmap contain a description of the resources, along with references to it's upstreams and downstreams.

Here's a sample API call:

```text
curl http://localhost:3000/v1/lineage?filter.type=bqtable

{
    "topic/events": {
        "urn": "events",
        "type": "topic",
        "downstreams": [
            "beast/events-ingestion"
        ],
        "upstreams": []
    },
    "beast/events-ingestion": {
        "urn": "events-ingestion",
        "type": "beast",
        "upstreams": [],
        "downstreams": [
            "bqtable/data-project:datalake.events"
        ]
    },
    "s3/events-transform-dwh": {
        "urn": "events-transform-dwh",
        "type": "s3",
        "upstreams": [
            "bqtable/data-project:datalake.events"
        ],
        "downstreams": [
            "bqtable/data-project:datawarehouse.events"
        ]
    },
    "bqtable/data-project:datalake.events": {
        "urn": "data-project:datalake.events",
        "type": "bqtable",
        "upstreams": [
            "beast/events-ingestion"
        ],
        "downstreams": [
            "s3/events-transform-dwh"
        ]
    },
    "bqtable/data-project:datawarehouse.events": {
        "urn": "data-project:datawarehouse.events",
        "type": "bqtable",
        "upstreams": [
            "s3/events-transform-dwh"
        ],
        "downstreams": []
    }
}
```

The node id's are are a concatentation of the resources "type" and it's "urn", separated by a `/`. This particular response depicts a data pipeline consisting of a "beast" application that persists data from a kafka topic to a bigquery table, and then a Optimus \(Bigquery Orchestrator\) job that processes and writes that data to a warehouse table.

Notice how we only queried for the `bqtable` lineage, yet the response contained resources of other types of resources that were related. This reflects how `bqtables` interfaces with other resources types. Anytime you query for a certain resources types, all resources types related to that resource, whether directly or indirectly are also returned.

But what if all you wished to know was how data flow's between two bigquery tables? Well, the Lineage API can optionally transform the requested lineage graph with just the dataflow information of the requested types using the `collapse` parameter.

by requesting the results to be collapse, the returned Lineage Graph is pre-processed by Compass to only contain the requested resources, and to mutate the references so that they point to the indirect ancestor/decendant that they're related to.

To demonstrate, let's make the same API call as above, but with `collapse` set to true.

```text
curl http://localhost:3000/v1/lineage?filter.type=bqtable&collapse=true


{
    "bqtable/data-project:datalake.events": {
        "urn": "data-project:datalake.events",
        "type": "bqtable",
        "upstreams": [],
        "downstreams": [
            "bqtable/data-project:datawarehouse.events"
        ]
    },
    "bqtable/data-project:datawarehouse.events": {
        "urn": "data-project:datawarehouse.events",
        "type": "bqtable",
        "upstreams": [
            "bqtable/data-project:datalake.events"
        ],
        "downstreams": []
    }
}
```

In this case, all other related resources and their references have been removed from the response. Additionally, notice how `bqtable/data-project:datalake.events` now declares `bqtable/data-project:datawarehouse.events` as it's downstream and vice versa.

Collapse can be used to request custom graphs of a specific subset of resources, ignoring any intermediate resource types that facilitate the data flow. For instance, to request the lineage graph containing just `beast` to `bqtable` dataflow, you can make the following API request:

```text
curl http://localhost:3000/v1/lineage?filter.type=bqtable&filter.type=beast&collapse=true
```

