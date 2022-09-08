package models

import (
	"log"
	"regexp"
)

type IndexStats struct {
	Health          string `json:"health" tsv:"health"`
	Status          string `json:"status" tsv:"status"`
	Index           string `json:"index" tsv:"index"`
	NoOfShards      int    `json:"no_of_shards" tsv:"pri"`
	NoOfReplica     int    `json:"no_of_replica" tsv:"rep"`
	DocCount        int    `json:"doc_count" tsv:"docs.count"`
	DeletedDocCount int    `json:"deleted_doc_count" tsv:"docs.deleted"`
	StorageSize     int    `json:"storage_size" tsv:"store.size"`
	PriStorageSize  int    `json:"pri_storage_size" tsv:"pri.store.size"`
}

type ShardStats struct {
	Index     string `json:"index" tsv:"index"`
	Shard     int    `json:"shard" tsv:"shard"`
	Type      string `json:"type" tsv:"prirep"`
	State     string `json:"state" tsv:"state"`
	Docs      int64  `json:"docs" tsv:"docs"`
	StoreSize int64  `json:"store_size" tsv:"store"`
	IpAddress string `json:"ip_address" tsv:"ip"`
	Node      string `json:"node" tsv:"node"`
}

func (ss *ShardStats) isPrimary() (primary bool) {
	if ss.Type == "p" {
		primary = true
	}
	return
}

func (ss *ShardStats) getIndexPattern() (pattern string) {
	reg, err := regexp.Compile("[0-9]")
	if err != nil {
		log.Fatal(err)
	}
	pattern = reg.ReplaceAllString(ss.Index, "*")
	return
}

type AllocationStats struct {
	Index       string `json:"index"`
	Shard       int    `json:"shard"`
	Type        string `json:"type"`
	IpAddress   string `json:"ip_address"`
	Segment     string `json:"segment"`
	Generation  int    `json:"generation"`
	Docs        int    `json:"docs"`
	DeletedDocs int    `json:"deleted_docs"`
	Size        int    `json:"size"`
	Searchable  bool   `json:"searchable"`
	Committed   bool   `json:"committed"`
	Version     string `json:"version"`
	Compound    string `json:"compound"`
}

type NodeStats struct {
	PrimaryShardsCount int    `json:"primary_shards_count"`
	ReplicaShardsCount int    `json:"replica_shards_count"`
	PrimarySizeBytes   int64  `json:"primary_size_bytes"`
	ReplicaSizeBytes   int64  `json:"replica_size_bytes"`
	NodeName           string `json:"node_name"`
}

func (node *NodeStats) adjustStats(status ShardStats) {
	if status.isPrimary() {
		node.PrimaryShardsCount++
		node.PrimarySizeBytes += status.StoreSize
	} else {
		node.ReplicaShardsCount++
		node.ReplicaSizeBytes += status.StoreSize
	}
}

type IndexNodeStats struct {
	Primaries int
	Replicas  int
}

func (in *IndexNodeStats) increasePrimary() {
	in.Primaries++
}

func (in *IndexNodeStats) increaseReplica() {
	in.Replicas++
}

func (in *IndexNodeStats) adjustShardCountFor(status ShardStats) {
	if status.isPrimary() {
		in.increasePrimary()
	} else {
		in.increaseReplica()
	}
}
