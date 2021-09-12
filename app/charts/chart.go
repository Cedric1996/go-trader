/*
 * @Author: cedric.jia
 * @Date: 2021-09-04 20:32:12
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-05 14:15:43
 */

package chart

import (
	"fmt"
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type Chart struct {
	filename string
}

func NewChart(filename string) *Chart {
	if len(filename) == 0 {
		filename = "bar"
	}
	return &Chart{filename: filename}
}

// generate random data for bar chart
// func (b *Chart) GenerateBarItems(datas []interface{}) []opts.BarData {
// 	items := make([]opts.BarData, 0)
// 	for _, data := range datas {
// 		items = append(items, opts.BarData{Value: data})
// 	}
// 	return items
// }

func BarCharts(xAxis []interface{}, series ...[]opts.BarData) *charts.Bar {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "vcp tr strategy",
		Subtitle: "bar chart for vcp_tr_strategy",
	}))

	// Put data into instance
	bar.SetXAxis(xAxis)
	for i, sery := range series {
		bar.AddSeries(fmt.Sprintf("Category %d", i), sery)
	}
	return bar
}

func ScatterCharts(xAxis []interface{}, series ...[]opts.ScatterData) *charts.Scatter {
	// create a new bar instance
	scatter := charts.NewScatter()
	// set some global options like Title/Legend/ToolTip or anything else
	scatter.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "vcp tr strategy",
		Subtitle: "bar chart for vcp_tr_strategy",
	}))

	// Put data into instance
	scatter.SetXAxis(xAxis)
	for i, sery := range series {
		scatter.AddSeries(fmt.Sprintf("Category %d", i), sery)
	}
	return scatter
}

func LineChart(xAxis []interface{},series ...[]opts.LineData) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "smooth style",
		}),
	)
	line.SetXAxis(xAxis)
	for i, sery := range series {
		line.AddSeries(fmt.Sprintf("Category %d", i), sery,charts.WithLineChartOpts(
			opts.LineChart{
				Smooth: true,
			}),
		)
	}
	return line
}


func (b *Chart) BarPage(charts ...components.Charter) {
	page := components.NewPage()
	page.AddCharts(charts...)

	// Where the magic happens
	f, err := os.Create(fmt.Sprintf(".charts/%s.html", b.filename))
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}
