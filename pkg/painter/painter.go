package painter

import (
    "fmt"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/plotutil"
    "gonum.org/v1/plot/vg"
)

type Chart struct {
    Title      string
    XLabel     string
    YLabel     string
    Width      vg.Length
    Height     vg.Length
    LinePoints []interface{}
    SvgFile    string
}

func DrawChart(chart *Chart) {
    p, err := plot.New()
    if err != nil {
        panic(err)
    }
    p.Title.Text = chart.Title
    p.X.Label.Text = chart.XLabel
    p.Y.Label.Text = chart.YLabel
    err = plotutil.AddLinePoints(p, chart.LinePoints...)
    if err != nil {
        panic(err)
    }
    if err := p.Save(chart.Width, chart.Height, chart.SvgFile); err != nil {
        panic(err)
    }
    fmt.Println("Successfully draw chart to ", chart.SvgFile)
}

func GetPointsFromSlice(slice []int) plotter.XYs {
    pts := make(plotter.XYs, len(slice))
    for i := range slice {
        pts[i].X = float64(i)
        pts[i].Y = float64(slice[i])
    }
    return pts
}

func GetPointsFromFloat64Slice(slice []float64) plotter.XYs {
    pts := make(plotter.XYs, len(slice))
    for i := range slice {
        pts[i].X = float64(i)
        pts[i].Y = slice[i]
    }
    return pts
}

func GetLinePoints(dataMap map[string][]int) []interface{} {
    var linePoints []interface{}
    for k,v := range dataMap {
        linePoints = append(linePoints, k, GetPointsFromSlice(v))
    }
    return linePoints
}