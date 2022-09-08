package config

import (
	"errors"
	"fmt"
	"shardanalyzer/models"
	"shardanalyzer/parser"
	"strings"
)

type ShardRecommendationRequest struct {
	CatShards         string
	TargetShardSizeGB int
	NumberOfAzs       int
	IsSearchWorkload  bool
	ClusterName       string
	ClientName        string
}

func (config *ShardRecommendationRequest) ParseStats() (cluster *models.Cluster, err error) {
	cluster = &models.Cluster{
		Name:                 config.ClusterName,
		ClientName:           config.ClientName,
		IsSearchWorkload:     config.IsSearchWorkload,
		NumberOfAZs:          config.NumberOfAzs,
		RecommendedShardSize: config.TargetShardSizeGB,
		Nodes:                map[string]*models.NodeStats{},
		Rollup:               map[string]*models.IndexPatternRollup{},
	}

	var contentReader = strings.NewReader(config.CatShards)
	shard := models.ShardStats{}
	var pars *parser.Parser
	valid := validateInput(config.CatShards)
	if !valid {
		return cluster, errors.New("input is not a valid one. Please provide the output of _cat/shards?v as input")
	}
	if strings.HasPrefix(config.CatShards, "index ") {
		pars, _ = parser.NewParser(contentReader, &shard)
	} else {
		pars = parser.NewParserWithoutHeader(contentReader, &shard)
	}
	for {
		eof, err := pars.Next()
		if eof {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue //ignoring if any missing information in shards line
		}
		cluster.Add(shard)
	}
	return
}

func validateInput(shards string) (valid bool) {
	if strings.HasPrefix(shards, "health status") {
		fmt.Println("Input looks like an output of _cat/indices")
		return
	}
	return true
}
