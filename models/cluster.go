package models

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const SingleIndexPattern = "--No Patterns--"

type Cluster struct {
	Name                  string
	ClientName            string
	NumberOfAZs           int
	IsSearchWorkload      bool
	RecommendedShardSize  int
	Rollup                map[string]*IndexPatternRollup
	Nodes                 map[string]*NodeStats
	TotalPrimarySizeBytes int64
	TotalReplicaSizeBytes int64
}

type Recommendation struct {
	Title                            string                       `json:"title"`
	ClusterName                      string                       `json:"cluster_name"`
	NumberOfAZs                      int                          `json:"number_of_azs"`
	NumberOfDataNodes                int                          `json:"number_of_data_nodes"`
	TotalPrimarySize                 int64                        `json:"total_primary_size"`
	TotalReplicaSize                 int64                        `json:"total_replica_size"`
	TotalShards                      int                          `json:"total_shards"`
	PotentialShards                  int                          `json:"potential_shards"`
	RecommendedShardSizeInGb         int                          `json:"recommended_shard_size_in_gb"`
	IndexPatternRecommendationRollup []IndexPatternRecommendation `json:"index_pattern_recommendation_rollup"`			// Array of IndexPatternRecommendation structs
	EmptyIndices                     []string                     `json:"empty_indices,omitempty"`						// Array of strings listing the empty indices
}

type IndexPatternRecommendation struct {
	Pattern                string                 `json:"pattern"`
	NeedChanges            bool                   `json:"need_changes"`
	Indices                []*IndexRecommendation `json:"indices"`														// Array of IndexRecommendation structs
	Size                   int64                  `json:"size"`
	PrimaryShards          int                    `json:"primary_shards"`
	ReplicaShards          int                    `json:"replica_shards"`
	PotentialPrimaryShards int                    `json:"potential_primary_shards"`
	PotentialReplicaShards int                    `json:"potential_replica_shards"`
	OldestIndexTime        time.Time              `json:"oldest_index_time"`
	NewestIndexTime        time.Time              `json:"newest_index_time"`
	Message                string                 `json:"message"`
	foundRotation          bool                   `json:"found_rotation"`
}

type IndexRecommendation struct {
	Name               string `json:"name"`
	Primaries          int    `json:"primaries"`
	PrimarySizeInBytes int64  `json:"primary_size_in_bytes"`
	PotentialPrimaries int    `json:"potential_primaries"`
	Docs               int64  `json:"docs"`
	Replicas           int    `json:"replicas"`
	PotentialReplicas  int    `json:"potential_replicas"`
}

func (c *Cluster) PrepareRecommendation() Recommendation {
	reco := Recommendation{
		Title:                            c.Name,
		ClusterName:                      c.ClientName,
		NumberOfAZs:                      c.NumberOfAZs,
		NumberOfDataNodes:                len(c.Nodes),
		RecommendedShardSizeInGb:         c.RecommendedShardSize,
		TotalPrimarySize:                 c.TotalPrimarySizeBytes,
		TotalReplicaSize:                 c.TotalReplicaSizeBytes,
		IndexPatternRecommendationRollup: []IndexPatternRecommendation{},
		EmptyIndices:                     []string{},
	}
	//var targetShardSizeInBytes int64
	targetShardSizeInBytes := int64(reco.RecommendedShardSizeInGb * 1024 * 1024 * 1024)
	sc := NewShardCounter(len(c.Nodes), c.NumberOfAZs)
	for pattern, ipr := range c.Rollup {
		//create pattern recommendation
		ipreco := IndexPatternRecommendation{
			Pattern: pattern,
			Indices: []*IndexRecommendation{},
		}
		for _, ir := range ipr.Indices {
			reco.TotalShards += ir.Replicas + ir.Primaries
			ireco := IndexRecommendation{
				Name:               ir.IndexName,
				PrimarySizeInBytes: ir.PrimarySizeBytes,
				Primaries:          ir.Primaries,
				Replicas:           ir.Replicas,
				Docs:               ir.Docs,
				PotentialReplicas:  ir.Replicas,
				PotentialPrimaries: ir.Primaries,
			}
			if ir.IsEmpty() {
				reco.AddEmptyIndex(ir)
				continue
			}

			ipreco.PrimaryShards += ir.Primaries
			ipreco.ReplicaShards += ir.Replicas
			ipreco.Size += ir.PrimarySizeBytes

			idealShardCount := sc.getIdealShardCount(ir.Primaries, ir.PrimarySizeBytes, targetShardSizeInBytes)
			if ir.Primaries != idealShardCount && idealShardCount > 0 {
				ipreco.NeedChanges = true
			}
			ireco.PotentialPrimaries = idealShardCount

			if c.IsSearchWorkload {
				ireco.PotentialReplicas = ireco.Replicas / ireco.Primaries
			} else {
				ireco.PotentialReplicas = 1
			}
			ipreco.PotentialPrimaryShards += ireco.PotentialPrimaries
			reco.PotentialShards += ireco.PotentialPrimaries
			if ireco.PotentialReplicas > 0 {
				replicaMultiplier := ireco.PotentialPrimaries * ireco.PotentialReplicas
				reco.PotentialShards += replicaMultiplier
				ipreco.PotentialReplicaShards += replicaMultiplier
			}
			ipreco.Indices = append(ipreco.Indices, &ireco)
		}
		//adjust replicas if there are potential warm indices
		ipreco.AdjustPotentialReplicaShards()
		//sort based in index names
		sort.Slice(ipreco.Indices, func(i, j int) bool {
			return ipreco.Indices[i].Name < ipreco.Indices[j].Name
		})
		reco.IndexPatternRecommendationRollup = append(reco.IndexPatternRecommendationRollup, ipreco)
	}

	return reco
}

func (c *Cluster) Add(status ShardStats) {
	pattern := status.Index

	if !c.IsSearchWorkload {
		//Don't rollup indices for search workloads
		pattern = status.getIndexPattern()
	}
	index := status.Index
	patternRollup := c.Rollup[pattern]
	if patternRollup == nil {
		//build one
		patternRollup = &IndexPatternRollup{Pattern: pattern, Parent: c, Indices: map[string]*IndexRollup{}}
		indexRollup := NewIndexRollup(index, patternRollup)
		indexRollup.add(status)
		patternRollup.Indices[status.Index] = indexRollup
		c.Rollup[pattern] = patternRollup
	} else {
		indexRollup := patternRollup.Indices[index]
		if indexRollup == nil {
			//create one
			indexRollup := NewIndexRollup(index, patternRollup)
			indexRollup.add(status)
			patternRollup.Indices[status.Index] = indexRollup
		} else {
			indexRollup.add(status)
		}
	}
	node := c.Nodes[status.Node]
	if node == nil {
		node = &NodeStats{
			NodeName: status.Node,
		}
		c.Nodes[status.Node] = node
	}
	node.adjustStats(status)
	if status.isPrimary() {
		c.TotalPrimarySizeBytes += status.StoreSize
	} else {
		c.TotalReplicaSizeBytes += status.StoreSize
	}
}

func (r Recommendation) GetIndexCount() (count int) {
	for _, ipr := range r.IndexPatternRecommendationRollup {
		count += len(ipr.Indices)
	}
	return
}

func (r *Recommendation) AddEmptyIndex(ir *IndexRollup) {
	if !strings.HasPrefix(ir.IndexName, ".") {
		r.EmptyIndices = append(r.EmptyIndices, ir.IndexName)
		//r.PotentialShards = r.PotentialShards - ir.Primaries - ir.Replicas
	}
}

func (r *Recommendation) NeedsShardAdjustment() bool {
	for _, ipr := range r.IndexPatternRecommendationRollup {
		if ipr.NeedChanges {
			return true
		}
	}
	return false
}

func (r *Recommendation) GetIndicesWithLargerShards(threshold int) (bool, []IndexRecommendation) {
	targetShardSizeInBytes := int64(threshold * 1024 * 1024 * 1024)
	var recommendation []IndexRecommendation
	//traverse all index patterns
	for _, ipr := range r.IndexPatternRecommendationRollup {
		if ipr.IsIndependentIndexPattern() {
			continue // as it was already processed with regular indices
		}
		for _, ir := range ipr.Indices {
			if ir.PrimarySizeInBytes/(int64(ir.Primaries)) > targetShardSizeInBytes {
				recommendation = append(recommendation, *ir)
			}
		}
	}
	return len(recommendation) > 0, recommendation
}

func (r Recommendation) GetTotalIndexPatterns() (total int) {

	for _, ipr := range r.IndexPatternRecommendationRollup {
		if !ipr.IsIndependentIndexPattern() {
			total++
		}
	}
	return
}

func (ipr *IndexPatternRecommendation) GetIndexDates() (oldest time.Time, newest time.Time, retentionInDays int) {
	//TODO: Implement to find oldest and newest index to find out the retention pattern
	return
}

func (ipr *IndexPatternRecommendation) GetCount() int {
	return len(ipr.Indices)
}

func (ipr *IndexPatternRecommendation) HasPotentialWarmIndices() (warmExist bool) {
	noReplicaCount := 0
	for _, ir := range ipr.Indices {
		if ir.Replicas == 0 {
			noReplicaCount++
		}
	}
	if noReplicaCount > 0 && ipr.GetCount() > noReplicaCount {
		// there are some indices with replica, it is fair to assume there are warm indices
		warmExist = true
	} else {
		// if all indices have no replicas, it is fair to assume the replica is missing.
		warmExist = false
	}
	return
}

func (ipr *IndexPatternRecommendation) AdjustPotentialReplicaShards() {
	if ipr.HasPotentialWarmIndices() {
		for _, ir := range ipr.Indices {
			if ir.Replicas == 0 {
				ipr.PotentialReplicaShards -= ir.PotentialReplicas
				ir.PotentialReplicas = 0
			}
		}
	}
}

func (ipr *IndexPatternRecommendation) IsIndependentIndexPattern() (independent bool) {
	if ipr.Pattern == SingleIndexPattern {
		independent = true
	}
	return
}

type IndexSettings struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

type IndexTemplate struct {
	IndexPatterns []string      `json:"index_patterns"`
	Settings      IndexSettings `json:"settings"`
}

func (ipr *IndexPatternRecommendation) GetIndexTemplateCommand() (cmd string) {
	it := IndexTemplate{
		IndexPatterns: []string{ipr.Pattern},
		Settings: IndexSettings{
			NumberOfReplicas: 1,
			NumberOfShards:   ipr.getRecommendedPrimaryShardsCount(),
		},
	}
	cmdBytes, err := PrettyStruct(it)
	if err == nil {
		cmd = "POST _template/" + strings.ReplaceAll(ipr.Pattern, "*", "") + "\n"
		return cmd + cmdBytes
	}
	return
}

func (ipr *IndexPatternRecommendation) getRecommendedPrimaryShardsCount() int {
	//var avgShardRecomended int
	dict := make(map[int]int)
	for _, indices := range ipr.Indices {
		dict[indices.PotentialPrimaries]++
	}
	common, shardCount := 0, 0
	for shard, count := range dict {
		if count > common {
			shardCount = shard
		}
	}
	//avgShardRecomended = ipr.PotentialPrimaryShards / len(ipr.Indices)
	//return avgShardRecomended
	return shardCount
}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}
