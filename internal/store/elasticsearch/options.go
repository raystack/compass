package elasticsearch

import "github.com/elastic/go-elasticsearch/v8"

type ClientOption func(*Client)

func WithClient(cli *elasticsearch.Client) ClientOption {
	return func(c *Client) {
		c.client = cli
	}
}
