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
				"suggester": {
					"type": "custom",
					"tokenizer": "non_alphanumeric",
					"filter": ["lowercase","preserve_original","shingle"]
				},
				"my_analyzer": {
					"type": "custom",
					"tokenizer": "my_tokenizer",
					"filter": ["lowercase","preserve_original","shingle"]
				}
			},
			"filter": {
				"shingle": {
					"type": "shingle",
					"min_shingle_size": 2,
					"max_shingle_size": 3
				},
				"preserve_original": {
      				"type": "word_delimiter",
					"preserve_original": "true"
				}
			},
			"tokenizer": {
			  "non_alphanumeric": {
				"type": "pattern",
				"pattern": "\\W|_"
			  },
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
				},
				"suggester": {
					"type": "text",
					"analyzer": "suggester"
				}
			}
		},
		"name": {
			"type": "text",
			"analyzer": "my_analyzer",
			"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256.0
				},
				"suggester": {
					"type": "text",
					"analyzer": "suggester"
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
