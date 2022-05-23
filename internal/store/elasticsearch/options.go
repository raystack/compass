package elasticsearch

import "github.com/elastic/go-elasticsearch/v7"

type ClientOption func(*Client)

func WithClient(cli *elasticsearch.Client) ClientOption {
	return func(c *Client) {
		c.client = cli
	}
}
