package template

import (
	"bytes"
	"text/template"

	jsoniter "github.com/json-iterator/go"
)

const costTemplate = `
window.costAnalysis = {{.}}
`

type AnalysisData struct {
	CostAnalysisByDay   CostAnalysis `json:"costAnalysisByDay"`
	CostAnalysisByMonth CostAnalysis `json:"costAnalysisByMonth"`
}

type CostAnalysis struct {
	ViewType   string             `json:"viewType"`
	DataCycle  string             `json:"dataCycle"`
	Statistics []ItemInStatistics `json:"statistics"`
	Ratios     []ItemInRatios     `json:"ratios"`
	CostTrend  *CostTrend         `json:"costTrend"`
}

type ItemInStatistics struct {
	SCycle     string `json:"sCycle"`
	SAmount    string `json:"sAmount"`
	SPreCycle  string `json:"sPreCycle"`
	SPreAmount string `json:"sPreAmount"`
	SRatio     string `json:"sRatio"`
}

type ItemInRatios struct {
	Chart ChartInRatios `json:"chart"`
}

type ChartInRatios struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	MidUnit  string            `json:"midUnit"`
	MidValue string            `json:"midValue"`
	Data     []ItemInRatioData `json:"data"`
}

type ItemInRatioData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CostTrend struct {
	Chart ChartInCostTrend `json:"chart"`
}
type TooltipUnit struct {
	Bar  string `json:"bar"`
	Line string `json:"line"`
}
type ChartInCostTrend struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	XData       []string       `json:"xData"`
	Series      []ItemInSeries `json:"series"`
	YTitle      []string       `json:"yTitle"`      // ["成本(元)", "变化比(%)"]
	TooltipUnit TooltipUnit    `json:"tooltipUnit"` // ["￥", "%"]
}

type ItemInSeries struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	YAxisIndex int      `json:"yAxisIndex"`
	Data       []string `json:"data"`
}

type costTemplateParams struct {
	CostAnalysis string
}

// ParseCostTemplate
func ParseCostTemplate(ad AnalysisData) (string, error) {
	s, _ := jsoniter.MarshalToString(ad)
	tmpl, _ := template.New("cost_template").Parse(costTemplate)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
