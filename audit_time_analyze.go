package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/iancoleman/orderedmap"
	"log"
	"os"
)

var result = orderedmap.New()

// 驳回的 actions
var rejectActions = map[string]bool{
	"OFF_SHELF":           true,
	"OFF_SHELF_INVISIBLE": true,
}

func filterByRejectActions(df dataframe.DataFrame, colName string) dataframe.DataFrame {
	actionCol := df.Col(colName)
	var filteredRows []int
	actionSeries := actionCol.Records()

	for i, value := range actionSeries {
		if rejectActions[value] {
			filteredRows = append(filteredRows, i)
		}
	}

	return df.Subset(filteredRows)
}

func calculateRatios(df dataframe.DataFrame, numFiles int, scenarioType string) {
	var filteredData dataframe.DataFrame
	if scenarioType == "闪拍/快捷开单" {
		filteredData = df.Filter(dataframe.F{Colname: "scenario_type", Comparator: "==", Comparando: scenarioType})
	} else {
		filteredData = df.Filter(dataframe.F{Colname: "scenario_type", Comparator: "==", Comparando: "普通"}).Filter(dataframe.F{Colname: "source", Comparator: "==", Comparando: scenarioType})
	}

	total := filteredData.Nrow()
	if total == 0 {
		return
	}

	machineAuditData := filteredData.Filter(dataframe.F{Colname: "audit_type", Comparator: "==", Comparando: "机审"})
	machineAuditRatio := float64(machineAuditData.Nrow()) / float64(total)
	machineRejectData := filterByRejectActions(machineAuditData, "action")
	machineRejectRatio := float64(machineRejectData.Nrow()) / float64(machineAuditData.Nrow())
	if machineAuditData.Nrow() == 0 {
		machineRejectRatio = 0
	}
	machineVerifiedRatio := 1 - machineRejectRatio
	machineMeanTime := machineAuditData.Col("audit_seconds").Mean()
	machineP90Time := machineAuditData.Col("audit_seconds").Quantile(0.90)
	machineP95Time := machineAuditData.Col("audit_seconds").Quantile(0.95)

	humanAuditData := filteredData.Filter(dataframe.F{Colname: "audit_type", Comparator: "==", Comparando: "人审"})
	humanData := orderedmap.New()
	if humanAuditData.Nrow() > 0 {
		humanRejectData := filterByRejectActions(humanAuditData, "action")
		humanRejectRatio := float64(humanRejectData.Nrow()) / float64(humanAuditData.Nrow())
		humanVerifiedRatio := 1 - humanRejectRatio
		humanMeanTime := humanAuditData.Col("audit_seconds").Mean()
		humanP90Time := humanAuditData.Col("audit_seconds").Quantile(0.90)
		humanP95Time := humanAuditData.Col("audit_seconds").Quantile(0.95)

		humanData.Set("人审量", fmt.Sprintf("%.2f 万", float64(humanAuditData.Nrow())/10000/float64(numFiles)))
		humanData.Set("人审驳回量", fmt.Sprintf("%.0f", float64(humanRejectData.Nrow())/float64(numFiles)))
		humanData.Set("人审通过率", fmt.Sprintf("%.2f%%", humanVerifiedRatio*100))
		humanData.Set("人审平均耗时", fmt.Sprintf("%.2f min", humanMeanTime/60))
		humanData.Set("人审P90耗时", fmt.Sprintf("%.2f min", humanP90Time/60))
		humanData.Set("人审P95耗时", fmt.Sprintf("%.2f min", humanP95Time/60))
	}

	meanTime := filteredData.Col("audit_seconds").Mean()
	p90Time := filteredData.Col("audit_seconds").Quantile(0.90)
	p95Time := filteredData.Col("audit_seconds").Quantile(0.95)

	scenarioResult := orderedmap.New()
	scenarioResult.Set("总数", fmt.Sprintf("%.2f 万", float64(total)/10000/float64(numFiles)))
	scenarioResult.Set("机审量", fmt.Sprintf("%.2f 万", float64(machineAuditData.Nrow())/10000/float64(numFiles)))
	scenarioResult.Set("机审率", fmt.Sprintf("%.2f%%", machineAuditRatio*100))
	scenarioResult.Set("机审驳回量", fmt.Sprintf("%.0f", float64(machineRejectData.Nrow())/float64(numFiles)))
	scenarioResult.Set("机审通过率", fmt.Sprintf("%.2f%%", machineVerifiedRatio*100))
	scenarioResult.Set("机审平均耗时", fmt.Sprintf("%.2f min", machineMeanTime/60))
	scenarioResult.Set("机审P90耗时", fmt.Sprintf("%.2f min", machineP90Time/60))
	scenarioResult.Set("机审P95耗时", fmt.Sprintf("%.2f min", machineP95Time/60))

	for _, key := range humanData.Keys() {
		val, _ := humanData.Get(key)
		scenarioResult.Set(key, val)
	}

	scenarioResult.Set("机审+人审平均耗时", fmt.Sprintf("%.2f min", meanTime/60))
	scenarioResult.Set("机审+人审P90耗时", fmt.Sprintf("%.2f min", p90Time/60))
	scenarioResult.Set("机审+人审P95耗时", fmt.Sprintf("%.2f min", p95Time/60))
	result.Set(scenarioType, scenarioResult)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: audit_time_analyze <file1> <file2> ... <fileN>")
		return
	}

	filePaths := os.Args[1:]
	var allDataFrames []dataframe.DataFrame

	for _, filePath := range filePaths {
		f, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}

		df := dataframe.ReadCSV(f, dataframe.WithTypes(map[string]series.Type{
			"is_channel": series.String,
		}))
		allDataFrames = append(allDataFrames, df)

		// 立即关闭文件以避免资源泄漏
		err = f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	// 合并所有 DataFrame
	var mergedDataFrame dataframe.DataFrame
	if len(allDataFrames) > 0 {
		mergedDataFrame = allDataFrames[0]
		for _, df := range allDataFrames[1:] {
			mergedDataFrame = mergedDataFrame.RBind(df)
		}
	}

	numFiles := len(filePaths)

	calculateRatios(mergedDataFrame, numFiles, "闪拍/快捷开单")
	calculateRatios(mergedDataFrame, numFiles, "开放平台")
	calculateRatios(mergedDataFrame, numFiles, "普通")

	total := mergedDataFrame.Nrow()
	machineAuditData := mergedDataFrame.Filter(dataframe.F{Colname: "audit_type", Comparator: "==", Comparando: "机审"})
	machineAuditRatio := float64(machineAuditData.Nrow()) / float64(total)

	allResult := orderedmap.New()
	allResult.Set("总数", fmt.Sprintf("%.2f 万", float64(total)/float64(numFiles)/10000))
	allResult.Set("机审量", fmt.Sprintf("%.2f 万", float64(machineAuditData.Nrow())/float64(numFiles)/10000))
	allResult.Set("机审率", fmt.Sprintf("%.2f%%", machineAuditRatio*100))
	result.Set("全部", allResult)

	jsonResult, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonResult))
}
