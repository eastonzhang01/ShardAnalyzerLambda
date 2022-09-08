package models

import "math"

type ShardCounter struct {
	DataNodes      int
	Azs            int
	PossibleCounts []int
}

func NewShardCounter(dataNodes int, azs int) *ShardCounter {
	sc := &ShardCounter{
		DataNodes:      dataNodes,
		Azs:            azs,
		PossibleCounts: getAllDivisibleNumbers(dataNodes, azs),
	}
	return sc
}

func (sc *ShardCounter) getIdealShardCount(primaries int, totalPrimarySize int64, targetPrimarySize int64) int {
	factor := float64(totalPrimarySize) / float64(targetPrimarySize)
	idealCount := int(math.Ceil(factor))
	if totalPrimarySize < targetPrimarySize {
		return 1
	} else {
		if idealCount == 1 {
			return 1
		}
		for _, v := range sc.PossibleCounts {
			if v >= idealCount {
				return v
			}
		}
	}
	// add the remainder if the idealCount is not evenly distributed.
	mod := idealCount % sc.DataNodes
	if mod == 0 {
		return idealCount
	} else {
		//closest divisible number
		for _, v := range sc.PossibleCounts {
			if v >= mod && primaries <= sc.DataNodes {
				return sc.DataNodes + v
			}
		}
		return idealCount + sc.DataNodes - mod
	}
}
