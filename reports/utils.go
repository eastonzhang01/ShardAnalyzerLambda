package reports

import (
	"fmt"
	"github.com/johnfercher/maroto/pkg/color"
	"shardanalyzer/models"
)

var (
	blue = color.Color{
		Red:   10,
		Green: 10,
		Blue:  150,
	}
	pacificSky = color.Color{
		Red:   185,
		Green: 217,
		Blue:  235,
	}

	sanFranciscoFog = color.Color{
		Red:   217,
		Green: 225,
		Blue:  226,
	}
	white = color.NewWhite()
)

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func MergeSingleIndexPatterns(recommendation *models.Recommendation) {
	indNRIndices := models.IndexPatternRecommendation{
		Pattern:     models.SingleIndexPattern,
		NeedChanges: false,
		Indices:     []*models.IndexRecommendation{},
	}
	indRIndices := models.IndexPatternRecommendation{
		Pattern:     models.SingleIndexPattern,
		NeedChanges: true,
		Indices:     []*models.IndexRecommendation{},
	}
	for _, ipr := range recommendation.IndexPatternRecommendationRollup {
		if len(ipr.Indices) == 1 {
			if ipr.NeedChanges {
				indRIndices.Indices = append(indRIndices.Indices, ipr.Indices...)
			} else {
				indNRIndices.Indices = append(indNRIndices.Indices, ipr.Indices...)
			}
		}
	}
	recommendation.IndexPatternRecommendationRollup = append(recommendation.IndexPatternRecommendationRollup, indRIndices, indNRIndices)
	return
}
