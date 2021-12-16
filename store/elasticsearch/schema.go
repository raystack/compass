package elasticsearch

// used as body to create index requests
// aliases the index to defaultSearchIndex
// and sets up the camelcase analyzer
var indexSettingsTemplate = `{
	"mappings": %s,
	"aliases": {
		%q: {}
	},
	"settings": {
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

var typeIndexMapping = `{
	"properties": {
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
		"description": {
			"type": "text"
		},
		"labels": {
			"type": "object"
		}
	}
}`
