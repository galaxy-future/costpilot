package template

import (
	"bytes"
	"text/template"

	jsoniter "github.com/json-iterator/go"
)

const utilizationTemplate = `
window.utilizeAnalysis = {{.}}
`

type ChartTrendStyle struct {
	Stack string `json:"stack"`
}

type UtilizeAnalysisItemInSeries struct {
	Name string   `json:"name"`
	Data []string `json:"data"`
}

type ChartCpuTrend struct {
	ID          string                        `json:"id"`
	Title       string                        `json:"title"`
	XData       []string                      `json:"xData"`
	Series      []UtilizeAnalysisItemInSeries `json:"series"`
	YTitle      []string                      `json:"yTitle"`
	TooltipUnit TooltipUnit                   `json:"tooltipUnit"`
	Style       ChartTrendStyle               `json:"style"`
	// Style = stack = total
}

type ChartUtilizeTrend struct {
	ID          string                        `json:"id"`
	Title       string                        `json:"title"`
	XData       []string                      `json:"xData"`
	Series      []UtilizeAnalysisItemInSeries `json:"series"`
	YTitle      []string                      `json:"yTitle"`
	TooltipUnit TooltipUnit                   `json:"tooltipUnit"`
}

type UtilizeAnalysisUtilizeTrend struct {
	Chart ChartUtilizeTrend `json:"chart"`
}

type UtilizeAnalysisCpuTrend struct {
	Chart ChartCpuTrend `json:"chart"`
}

type UtilizeAnalysisStatisticsItem struct {
	SCycle     string `json:"sCycle"`
	SAmount    string `json:"sAmount"`
	SUnit      string `json:"sUnit"`
	SPreCycle  string `json:"sPreCycle"`
	SPreAmount string `json:"sPreAmount"`
	SPreUnit   string `json:"sPreUnit"`
	SRatio     string `json:"sRatio"`
}

type UtilizeAnalysis struct {
	AnalysisByDay UtilizeAnalysisByDay `json:"utilizeAnalysisByDay"`
}
type UtilizeAnalysisByDay struct {
	ViewType     string                          `json:"viewType"`
	DataCycle    string                          `json:"dataCycle"`
	Statistics   []UtilizeAnalysisStatisticsItem `json:"statistics"`
	Ratios       []ItemInRatios                  `json:"ratios"`
	UtilizeTrend *UtilizeAnalysisUtilizeTrend    `json:"utilizeTrend"`
	CpuTrend     *UtilizeAnalysisCpuTrend        `json:"cpuTrend"`
}

func ParseUtilizeAnalysisTemplate(ua UtilizeAnalysis) (string, error) {
	s, _ := jsoniter.MarshalToString(ua)
	tmpl, _ := template.New("utilize_analysis_template").Parse(utilizationTemplate)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
