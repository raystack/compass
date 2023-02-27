package elasticsearch

// used as body to create index requests
// aliases the index to defaultSearchIndexAlias
// and sets up the camelcase analyzer
var indexSettingsTemplate = `{
	"mappings": %s,
	"aliases": {
		%q: {}
	},
	"settings": {
		"index": {
            "number_of_shards": %d
        },
		"analysis": {
			"analyzer": {
				"my_analyzer": {
					"type": "custom",
					"tokenizer": "my_tokenizer",
					"filter": ["lowercase"]
				}
			},
			"tokenizer": {
			  "my_tokenizer": {
				"type": "pattern",
				"pattern": "([^\\p{L}\\d]+)|(?<=\\D)(?=\\d)|(?<=\\d)(?=\\D)|(?<=[\\p{L}&&[^\\p{Lu}]])(?=\\p{Lu})|(?<=\\p{Lu})(?=\\p{Lu}[\\p{L}&&[^\\p{Lu}]])"
			  }
			}
		}
	}
}`

var serviceIndexMapping = `{
	"properties": {
		"namespace_id": {
			"type": "keyword"
		},
		"urn": {
			"type": "text",
			"analyzer": "my_analyzer",
			"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256.0
				}
			}
		},
		"name": {
			"type": "text",
			"analyzer": "my_analyzer",
			"fields": {
				"suggest": {
					"type": "completion"
				},
				"keyword": {
					"type": "keyword",
					"ignore_above": 256.0
				}
			}
		},
		"service": {
			"type": "text",
			"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256.0
				}
			}
		},
		"type": {
			"type": "text",
			"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256.0
				}
			}
		},
		"description": {
			"type": "text"
		},
		"labels": {
			"type": "object"
		}
	}
}`

// indexSharedAliasSettingsTemplate is used as body to create index alias requests for shared tenants
// item read/write calls will be routed to specific shards and then filtered before fetching by tenant id
var indexSharedAliasSettingsTemplate = `{
    "actions": [
        {
            "remove": {
                "alias": "{{alias_name}}",
                "index": "{{index_name}}"
            }
        },
        {
            "add": {
                "index": "{{index_name}}",
                "alias": "{{alias_name}}",
                "filter": {
                    "term": {
                        "namespace_id": "{{filter_id}}"
                    }
                },
                "index_routing": "{{write_id}}",
                "search_routing": "{{read_id}}"
            }
        }
    ]
}`

// indexDedicatedAliasSettingsTemplate is used as body to create index alias requests for dedicated tenants
// for dedicated tenants, it's not required to route read/write and filter items based on namespace
// as the whole index belongs to tenant
var indexDedicatedAliasSettingsTemplate = `{
    "actions": [
        {
            "remove": {
                "alias": "{{alias_name}}",
                "index": "{{index_name}}"
            }
        },
        {
            "add": {
                "index": "{{index_name}}",
                "alias": "{{alias_name}}"
            }
        }
    ]
}`
