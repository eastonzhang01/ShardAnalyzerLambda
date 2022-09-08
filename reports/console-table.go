package reports

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"shardanalyzer/models"
	"strconv"
)

func RenderAsTable(recommendation models.Recommendation) string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Index Name", "Primary Size", "Current Settings", "Recommended Settings"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold, tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})

	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor})

	noChangesNeeded := []tablewriter.Colors{{tablewriter.Bold},
		{tablewriter.Bold},
		{tablewriter.Bold, tablewriter.FgGreenColor},
		{tablewriter.Bold, tablewriter.FgHiGreenColor}}
	for _, ipr := range recommendation.IndexPatternRecommendationRollup {
		if ipr.NeedChanges {
			if ipr.IsIndependentIndexPattern() || len(ipr.Indices) > 1 {
				patternData := []string{"IndexPattern=" + ipr.Pattern}
				table.Rich(patternData, []tablewriter.Colors{{tablewriter.Normal, tablewriter.ALIGN_CENTER, tablewriter.BgGreenColor, tablewriter.FgBlackColor}, {tablewriter.Normal, tablewriter.BgGreenColor, tablewriter.FgBlackColor}, {tablewriter.Normal, tablewriter.BgGreenColor, tablewriter.FgBlackColor}})
				for _, ir := range ipr.Indices {
					rowData := []string{ir.Name, ByteCountIEC(ir.PrimarySizeInBytes), strconv.Itoa(ir.Primaries) + "/" + strconv.Itoa(ir.Replicas/ir.Primaries), strconv.Itoa(ir.PotentialPrimaries) + "/" + strconv.Itoa(ir.PotentialReplicas)}
					if ir.Primaries == ir.PotentialPrimaries && ir.Replicas == ir.PotentialReplicas {
						table.Rich(rowData, noChangesNeeded)
					} else {
						table.Append(rowData)
					}
				}
			}
		}

	}

	table.SetFooter([]string{"", "", strconv.Itoa(recommendation.TotalShards), strconv.Itoa(recommendation.PotentialShards)}) // Add Footer
	//table.SetRowLine(true)
	//table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	//table.SetCenterSeparator("|")
	//table.SetAutoMergeCells(true)
	table.SetAutoFormatHeaders(true)
	table.Render()
	return buf.String()
}
