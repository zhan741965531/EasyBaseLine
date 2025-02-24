package report

import (
	"EasyBaseLine/config"
	"EasyBaseLine/utils"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/xuri/excelize/v2"
	"time"
)

var infoColor = color.New(color.FgHiGreen).SprintFunc()

func GenerateReport(finalResult config.FinalResult) {
	results := finalResult.CheckResults

	file := excelize.NewFile()
	prepareSheet(file)

	styles := createStyles(file)
	fillSheet(file, results, styles)

	saveFile(file, finalResult.HostIP, finalResult.BasicInfo)
}

func prepareSheet(file *excelize.File) {
	sheetName := "Report"
	_, err := file.NewSheet(sheetName)
	if err != nil {
		utils.Fatal(fmt.Sprintf("创建工作表失败: %v\n", err), color.New(color.FgHiRed).SprintFunc())
		return
	}

	colWidths := []float64{7.11, 60, 8.78, 4.78, 116, 81, 154}
	for i, w := range colWidths {
		file.SetColWidth(sheetName, string('A'+i), string('A'+i), w)
	}
}

type Styles struct {
	Normal int
	Header int
	Pass   int
	Fail   int
	Check  int
}

func createStyles(file *excelize.File) Styles {
	borderFull := `{"border":[{"type":"top","color":"#000000","style":1},{"type":"bottom","color":"#000000","style":1},{"type":"left","color":"#000000","style":1},{"type":"right","color":"#000000","style":1}]}`

	return Styles{
		Normal: createStyle(file, borderFull),
		Header: createStyle(file, mergeStyles(borderFull, `{"font":{"bold":true},"alignment":{"horizontal":"center"},"fill":{"type":"pattern","color":["#c3d1e6"],"pattern":1}}`)),
		Pass:   createStyle(file, mergeStyles(borderFull, `{"fill":{"type":"pattern","color":["#c6efce"],"pattern":1}}`)),
		Fail:   createStyle(file, mergeStyles(borderFull, `{"fill":{"type":"pattern","color":["#ffc7ce"],"pattern":1}}`)),
		Check:  createStyle(file, mergeStyles(borderFull, `{"fill":{"type":"pattern","color":["#ffeb9c"],"pattern":1}}`)),
	}
}

func fillSheet(file *excelize.File, results []config.CheckResult, styles Styles) {
	sheetName := "Report"
	setCellValues(file, sheetName, styles.Header,
		"A1", "UID",
		"B1", "描述",
		"C1", "风险等级",
		"D1", "状态",
		"E1", "输出",
		"F1", "危害",
		"G1", "解决方案",
	)

	for index, result := range results {
		row := index + 2
		setCellValues(file, sheetName, styles.Normal,
			fmt.Sprintf("A%d", row), result.UID,
			fmt.Sprintf("B%d", row), result.Description,
			fmt.Sprintf("C%d", row), result.RiskLevel,
			fmt.Sprintf("D%d", row), result.Status,
			fmt.Sprintf("E%d", row), result.OutPuts,
			fmt.Sprintf("F%d", row), result.Harm,
			fmt.Sprintf("G%d", row), result.Solution,
		)

		statusCell := fmt.Sprintf("D%d", row)
		switch result.Status {
		case "通过":
			file.SetCellStyle(sheetName, statusCell, statusCell, styles.Pass)
		case "失败":
			file.SetCellStyle(sheetName, statusCell, statusCell, styles.Fail)
		case "人工检查":
			file.SetCellStyle(sheetName, statusCell, statusCell, styles.Check)
		}
	}
}

func setCellValues(file *excelize.File, sheetName string, style int, cells ...string) {
	for i := 0; i < len(cells); i += 2 {
		cell, value := cells[i], cells[i+1]
		if err := file.SetCellValue(sheetName, cell, value); err != nil {
			utils.Fail(fmt.Sprintf("设置单元格值时出错: %v", err), color.New(color.FgHiGreen).SprintFunc())
			continue
		}
		file.SetCellStyle(sheetName, cell, cell, style)
	}
}

func createStyle(file *excelize.File, styleJSON string) int {
	var style excelize.Style
	if err := json.Unmarshal([]byte(styleJSON), &style); err != nil {
		utils.Fatal(fmt.Sprintf("解析样式失败: %v\n", err), color.New(color.FgHiRed).SprintFunc())
		return -1
	}

	styleIndex, err := file.NewStyle(&style)
	if err != nil {
		utils.Fatal(fmt.Sprintf("创建新的样式时出错: %v\n", err), color.New(color.FgHiRed).SprintFunc())
		return -1
	}

	return styleIndex
}

func saveFile(file *excelize.File, ip string, basicInfo config.BasicInfo) {
	if ip == "" {
		var err error
		ip, err = utils.GetIPAddress()
		if err != nil {
			utils.Warn(fmt.Sprintf("获取IP地址出现错误: %s，将使用默认IP: 1.2.3.4\n", err), color.New(color.FgHiYellow).SprintFunc())
			ip = "1.2.3.4"
		}
	}

	fileName := fmt.Sprintf("%s_%s.xlsx", time.Now().Format("2006-01-02-15-04-05"), ip)
	if err := file.SaveAs(fileName); err != nil {
		utils.Warn(fmt.Sprintf("保存文件 %s 时出错: %v\n", fileName, err), color.New(color.FgHiRed).SprintFunc())
		return
	}

	utils.Info(fmt.Sprintf("[%s] 检查报告已生成: %v\n", infoColor(basicInfo.CheckName), infoColor(fileName)))
}

func mergeStyles(styles ...string) string {
	mergedStyle := make(map[string]interface{})
	for _, style := range styles {
		var styleDict map[string]interface{}
		if err := json.Unmarshal([]byte(style), &styleDict); err != nil {
			utils.Warn(fmt.Sprintf("解析样式字符串时出错: %v\n", err), color.New(color.FgHiYellow).SprintFunc())
			continue
		}
		for key, value := range styleDict {
			mergedStyle[key] = value
		}
	}

	mergedStyleBytes, err := json.Marshal(mergedStyle)
	if err != nil {
		utils.Warn(fmt.Sprintf("转换合并样式时出错: %v\n", err), color.New(color.FgHiYellow).SprintFunc())
		return "{}"
	}

	return string(mergedStyleBytes)
}
