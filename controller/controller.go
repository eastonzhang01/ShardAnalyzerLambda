package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"shardanalyzer/config"
	"shardanalyzer/reports"
	"strconv"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Simple group: v1
	v1 := r.Group("/v1")
	{
		shardAnalyzerGroup := v1.Group("/shard-analyzer")
		{
			shardAnalyzerGroup.POST("", recommend)
		}

	}

	return r
}

// @Summary Recommend shard strategies
// @Description Endpoint to take cat/shards output and examine they are rightly sharded or not and recommend the right ones if necessary.
// @Id shardAnalyzerPost
// @Accept application/text
// @Produce application/json, application/pdf
// @Param clusterName query string true "Cluster Name" default(Cluster-name)
// @Param customerName query string true "Customer Name" default(AWS Customer)
// @Param targetShardSize query int true "Target Shard Size in GB" default(30)
// @Param azs query int true "Number of Azs for the cluster" default(3)
// @Param isSearchWorkload query bool false "If log analytics, 1 replica is recommended. If not the replica count will be retained" default(false)
// @Param query body string true "Output of cat/shards."
// @Success 400 {string} string
// @Failure 500 {string} string
// @Router /v1/shard-analyzer [post]
func recommend(context *gin.Context) {
	body, err := context.GetRawData()
	if err != nil {
		context.String(http.StatusBadRequest, "Error in getting request body")
		return
	}
	targetstr, _ := context.GetQuery("targetShardSize")
	targetSize, err := strconv.Atoi(targetstr)
	if err != nil {
		context.String(http.StatusBadRequest, "targetShardSize must be an integer")
		return
	}
	azstr, _ := context.GetQuery("azs")
	numOfAzs, err := strconv.Atoi(azstr)
	if err != nil || numOfAzs > 3 || numOfAzs < 1 {
		context.String(http.StatusBadRequest, "azs must be an integer between 1 to 3")
		return
	}
	isSearchStr, _ := context.GetQuery("isSearchWorkload")
	isSearch, err := strconv.ParseBool(isSearchStr)
	if err != nil {
		context.String(http.StatusBadRequest, "isLogAnalytics must be an either true or false")
		return
	}
	clusterName, _ := context.GetQuery("clusterName")
																		// All above parses through post request fills inputs with what was passed in 

	args := config.ShardRecommendationRequest{							// Create Struct of all the user inputs
		CatShards:         string(body),
		TargetShardSizeGB: targetSize,
		NumberOfAzs:       numOfAzs,
		IsSearchWorkload:  isSearch,
		ClusterName:       clusterName,
	}
	cluster, err := args.ParseStats()									// Parse through inputs given
	if err != nil {
		context.JSON(http.StatusBadRequest, err)
		fmt.Println("Can't proceed. Error: ", err)
		return
	}
	println("total shards", len(cluster.Rollup))

	recommendation := cluster.PrepareRecommendation()					// Generate JSON of recommendation

	reports.MergeSingleIndexPatterns(&recommendation)					// Not sure what this does
	//tableString := reports.RenderAsTable(recommendation)

	//buf, err:=reports.GeneratePDFResponse(recommendation)
	//if err != nil {
	//	context.String(http.StatusInternalServerError, "error while creating PDF report")
	//} else {
	//	context.Writer.Header().Set("Content-type", "application/pdf")
	//	context.Writer.Write(buf.Bytes())
	//}
	context.JSON(http.StatusOK, recommendation)							// Not sure what this does
}
