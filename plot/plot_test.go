package plot

import (
	"gonum.org/v1/plot/plotter"
	"testing"
)

func TestPlot(t *testing.T) {
	XYs := make([]interface{}, 0)
	XYsi := plotter.XYs{{0.1, 1}, {0.3, 0.5}}

	XYs = append(XYs, `tesssss`, XYsi)
	Save(XYs, `test`, 1)
}
