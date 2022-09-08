package reports

import (
	"bytes"
	"fmt"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"os"
	"shardanalyzer/config"
	"shardanalyzer/models"
	"strconv"
	"time"
)

func GeneratePDFReport(recommendation models.Recommendation, nodes map[string]*models.NodeStats, isClusterWise bool, fileName string) {
	begin := time.Now()
	// m is the PDF report 
	m := preparePDFReport(recommendation, nodes, isClusterWise)

	err := m.OutputFileAndClose(fileName)																								
	if err != nil {
		fmt.Println("File not saved to tmp", err)
		os.Exit(1)
	}

	end := time.Now()
	fmt.Println("Total time took to generate report ", end.Sub(begin))
	/*
	open, err := os.Open(fileName)																										
	fmt.Println("Report available at ", open.Name())
	defer func(open *os.File) {
		err := open.Close()
		if err != nil {
			fmt.Println("error in closing file", err)
		}
	}(open)
	*/
}

func GeneratePDFResponse(recommendation models.Recommendation, nodes map[string]*models.NodeStats) (buf bytes.Buffer, err error) {
	begin := time.Now()
	m := preparePDFReport(recommendation, nodes, false)

	buf, err = m.Output()
	if err != nil {
		fmt.Println("Could not save PDF:", err)
		return
	}

	end := time.Now()
	fmt.Println(end.Sub(begin))
	return
}

func preparePDFReport(recommendation models.Recommendation, nodes map[string]*models.NodeStats, isClusterWise bool) pdf.Maroto {
	header := getHeader()

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(10, 15, 10)
	m.SetAliasNbPages("{nb}")
	m.SetFirstPageNb(1)
	addHeaderFooter(m, blue)

	m.Row(7, func() {
		m.Col(12, func() {
			m.Text("Cluster report for "+recommendation.ClusterName, props.Text{
				Top:    3,
				Family: consts.Helvetica,
				Style:  consts.Bold,
				Align:  consts.Center,
			})
		})
	})

	m.Row(2, func() {
		m.Col(12, func() {
			m.Text(time.Now().Format("Mon Jan 2, 2006"), props.Text{
				Top:    1.5,
				Size:   6,
				Family: consts.Helvetica,
				Style:  consts.Bold,
				Align:  consts.Center,
			})
		})
	})
	m.Row(5, func() {})

	addHeader("Cluster Details", m)

	m.TableList([]string{"Attribute", "Value"}, getClusterAttributes(recommendation), getTwoColumnLeftAlignedTableList(sanFranciscoFog))

	if !isClusterWise {
		addHeader("Shard Recommendations for indices", m)
		m.Row(3, func() {})
		m.TableList(header, getContents(recommendation), getIndexTableList(sanFranciscoFog))
	} else {
		addLargerIndices(recommendation, m)
		for _, ipr := range recommendation.IndexPatternRecommendationRollup {
			if ipr.NeedChanges {
				var data [][]string
				data = getRowContent(ipr)
				if len(data) > 0 {
					addHeader("Recommendation for Index pattern "+ipr.Pattern, m)
					if !ipr.IsIndependentIndexPattern() {
						m.TableList([]string{"Attribute", "Value"}, getPatternAttributes(ipr), getTwoColumnLeftAlignedTableList(sanFranciscoFog))
					}
					m.Row(4, func() {})
					m.TableList(header, data, getIndexTableList(sanFranciscoFog))
				}
			}
		}
	}
	//add cluster skew analysis
	addClusterSkewAnalysis(nodes, m)
	return m
}

func getBuildStr() string {
	if config.Version == "" {
		config.Version = "1.0.0"
		config.Build = "Development"
	}
	by := fmt.Sprintf("Generated with ShardAnalyzer version: %s , build: %s", config.Version, config.Build)
	return by
}

func addClusterSkewAnalysis(nodes map[string]*models.NodeStats, m pdf.Maroto) {
	addHeader("Node distribution", m)
	m.TableList([]string{"Node", "Total shards (P/R)", "Size"}, getNodeDetails(nodes), getClusterTableList(sanFranciscoFog))
}

func getNodeDetails(nodes map[string]*models.NodeStats) (data [][]string) {
	for _, ns := range nodes {
		total := strconv.Itoa(ns.PrimaryShardsCount+ns.ReplicaShardsCount) + " (" + strconv.Itoa(ns.PrimaryShardsCount) + "/" + strconv.Itoa(ns.ReplicaShardsCount) + ")"
		size := ByteCountIEC(ns.PrimarySizeBytes+ns.ReplicaSizeBytes) + " (" + ByteCountIEC(ns.PrimarySizeBytes) + "/" + ByteCountIEC(ns.ReplicaSizeBytes) + ")"
		data = append(data, []string{ns.NodeName, total, size})
	}
	return
}

func addLargerIndices(recommendation models.Recommendation, m pdf.Maroto) {
	available, indices := recommendation.GetIndicesWithLargerShards(50)
	if available {
		var data [][]string
		for _, ir := range indices {
			rowData := []string{ir.Name, ByteCountIEC(ir.PrimarySizeInBytes), strconv.Itoa(ir.Primaries) + "/" + strconv.Itoa(ir.Replicas/ir.Primaries), strconv.Itoa(ir.PotentialPrimaries) + "/" + strconv.Itoa(ir.PotentialReplicas)}
			data = append(data, rowData)
		}
		addHeader("Indices have shards greater than 50GB", m)
		m.TableList(getHeader(), data, getIndexTableList(sanFranciscoFog))
	}
}

func addHeader(header string, m pdf.Maroto) {
	m.SetBackgroundColor(white)
	m.TableList([]string{""}, [][]string{{header}}, getBoxTableList(pacificSky))
	m.Row(3, func() {})
}

func getPatternAttributes(ipr models.IndexPatternRecommendation) (data [][]string) {
	data = [][]string{
		{"Pattern Name", ipr.Pattern},
		{"No of indices found in this pattern", strconv.Itoa(ipr.GetCount())},
		{"Primary Shards", strconv.Itoa(ipr.PrimaryShards)},
		{"Replica Shards", strconv.Itoa(ipr.ReplicaShards)},
		{"Size of Primary Indices", ByteCountIEC(ipr.Size)},
		{"Potential Primary Shards", strconv.Itoa(ipr.PotentialPrimaryShards)},
		{"Potential Replica Shards", strconv.Itoa(ipr.PotentialReplicaShards)},
		//
		{"Recommended Index template", ipr.GetIndexTemplateCommand()},
	}
	fmt.Println(ipr.GetIndexTemplateCommand())
	return
}

func getTwoColumnLeftAlignedTableList(color color.Color) props.TableList {
	return props.TableList{
		HeaderProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{3, 9},
			Family:    consts.Helvetica,
		},
		ContentProp: props.TableListContent{
			Size:      8,
			GridSizes: []uint{3, 9},
			Family:    consts.Helvetica,
		},
		Align:                consts.Left,
		AlternatedBackground: &color,
		HeaderContentSpace:   1,
		Line:                 false,
	}
}

func getBoxTableList(grayColor color.Color) props.TableList {
	return props.TableList{
		ContentProp: props.TableListContent{
			Size:      8,
			GridSizes: []uint{12},
			Family:    consts.Helvetica,
			Style:     consts.Bold,
		},
		Align:                consts.Center,
		AlternatedBackground: &grayColor,
		HeaderContentSpace:   0,
		Line:                 false,
	}
}

func getIndexTableList(color color.Color) props.TableList {
	return props.TableList{
		HeaderProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{6, 2, 2, 2},
			Family:    consts.Helvetica,
		},
		ContentProp: props.TableListContent{
			Size:      8,
			GridSizes: []uint{6, 2, 2, 2},
			Family:    consts.Helvetica,
		},
		Align:                consts.Center,
		AlternatedBackground: &color,
		HeaderContentSpace:   1,
		Line:                 false,
	}
}

func getClusterTableList(color color.Color) props.TableList {
	return props.TableList{
		HeaderProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{3, 3, 2, 2, 2},
			Family:    consts.Helvetica,
		},
		ContentProp: props.TableListContent{
			Size:      8,
			GridSizes: []uint{3, 3, 2, 2, 2},
			Family:    consts.Helvetica,
		},
		Align:                consts.Center,
		AlternatedBackground: &color,
		HeaderContentSpace:   1,
		Line:                 false,
	}
}

func addHeaderFooter(m pdf.Maroto, color color.Color) {
	m.RegisterHeader(func() {
		m.Row(3, func() {
			m.Col(12, func() {
				m.Text(getBuildStr(), props.Text{
					Style:  consts.Normal,
					Size:   6,
					Family: consts.Helvetica,
					Align:  consts.Right,
				})
			})
		})
	})
	m.RegisterFooter(func() {
		m.Row(6, func() {
			m.Col(5, func() {
				m.Text("aws.amazon.com/opensearch-service", props.Text{
					Style: consts.BoldItalic,
					Size:  8,
					Align: consts.Left,
					Color: color,
				})
			})
			m.Col(6, func() {
				m.Text(strconv.Itoa(m.GetCurrentPage())+"/{nb}", props.Text{
					Align: consts.Right,
					Size:  8,
					Color: color,
				})
			})
		})
	})
}

func getClusterAttributes(recommendation models.Recommendation) (data [][]string) {
	data = [][]string{
		{"Cluster Name", recommendation.Title},
		{"Number of Data nodes", strconv.Itoa(recommendation.NumberOfDataNodes)},
		{"Number of AZs", strconv.Itoa(recommendation.NumberOfAZs)},
		{"Total Indices", strconv.Itoa(recommendation.GetIndexCount())},					// Index Count is from summing length of each Indices array within the IndexPatternRecommendation that is within the IndexPatternRecommendationRollup array
		{"Total index patterns", strconv.Itoa(recommendation.GetTotalIndexPatterns())},		// number of IndexPatternRecommendation structs that have Pattern!=No Patterns
		{"Size of Primary Indices", ByteCountIEC(recommendation.TotalPrimarySize)},
		{"Size of Replica Indices", ByteCountIEC(recommendation.TotalReplicaSize)},
		{"Target Shard Size in GB", strconv.Itoa(recommendation.RecommendedShardSizeInGb)},
		{"Total Shards", strconv.Itoa(recommendation.TotalShards)},
		{"Potential Shards", strconv.Itoa(recommendation.PotentialShards)},
		
		// is missing recommendation.ClusterName
	}
	if len(recommendation.EmptyIndices) > 0 {
		data = append(data, []string{
			"Empty Indices", fmt.Sprint(recommendation.EmptyIndices),
		})
	}
	return
}

func getHeader() []string {
	return []string{"Index Name", "Primary Size", "Current Settings p/r", "Recommended Settings p/r"}
}

func getContents(recommendation models.Recommendation) (data [][]string) {
	data = [][]string{}
	for _, ipr := range recommendation.IndexPatternRecommendationRollup {
		if ipr.NeedChanges {
			data = getRowContent(ipr)
		}

	}
	return
}

func getRowContent(ipr models.IndexPatternRecommendation) (data [][]string) {
	if ipr.IsIndependentIndexPattern() || len(ipr.Indices) > 1 {
		for _, ir := range ipr.Indices {
			rowData := []string{ir.Name, ByteCountIEC(ir.PrimarySizeInBytes), strconv.Itoa(ir.Primaries) + "/" + strconv.Itoa(ir.Replicas/ir.Primaries), strconv.Itoa(ir.PotentialPrimaries) + "/" + strconv.Itoa(ir.PotentialReplicas)}
			data = append(data, rowData)
		}
	}
	return data
}
