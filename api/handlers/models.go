package handlers

import (
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lineage"
)

type upsertAssertPayload struct {
	asset.Asset
	Upstreams   []lineage.Node `json:"upstreams"`
	Downstreams []lineage.Node `json:"downstreams"`
}
