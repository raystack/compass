package lineage

// dataflowDir describes the direction of data IO with regards to another resource
// it's "upstream" if the application only reads, "downstream" if the application only writes
// "bidirectional" if the application both reads and writes to another resource.
// "bidirectional" direction is currently not supported.
type dataflowDir string

const (
	dataflowDirUpstream   = dataflowDir("upstream")
	dataflowDirDownstream = dataflowDir("downstream")
)
