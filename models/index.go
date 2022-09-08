package models

import (
	"strconv"
)

type IndexPatternRollup struct {
	Pattern string
	Parent  *Cluster
	Indices map[string]*IndexRollup
}

type IndexRollup struct {
	IndexName        string
	PrimarySizeBytes int64
	ReplicaSizeBytes int64
	Primaries        int
	Replicas         int
	Shards           map[string]*ShardStats
	Docs             int64
	Nodes            map[string]*IndexNodeStats
	Parent           *IndexPatternRollup
}

func (ir *IndexRollup) add(status ShardStats) {
	shardName := status.Index + "[" + strconv.Itoa(status.Shard) + "]" + status.Type
	ir.Shards[shardName] = &status
	ir.Docs += status.Docs
	indexNodeStats := ir.Nodes[status.Node]
	if indexNodeStats == nil {
		indexNodeStats = &IndexNodeStats{}
		ir.Nodes[status.Node] = indexNodeStats
	}
	indexNodeStats.adjustShardCountFor(status)
	if status.isPrimary() {
		ir.PrimarySizeBytes += status.StoreSize
		ir.Primaries++
	} else {
		ir.ReplicaSizeBytes += status.StoreSize
		ir.Replicas++
	}
}

func (ir *IndexRollup) IsEmpty() bool {
	return ir.Docs <= 0
}

func (ir *IndexRollup) IsPotentialUWIndex() bool {
	return ir.Replicas == 0
}

func NewIndexRollup(name string, pattern *IndexPatternRollup) *IndexRollup {
	return &IndexRollup{
		IndexName: name,
		Parent:    pattern,
		Shards:    map[string]*ShardStats{},
		Nodes:     map[string]*IndexNodeStats{},
	}
}
