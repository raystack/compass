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
				"default": {
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
			"type": "text"
		},
		"name": {
			"type": "text"
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
