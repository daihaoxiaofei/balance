// 用于输出折线图 回测时比较直观
package plot

import (
	"log"
	"math/rand"
	"os"
	"path"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const Path = `png`

func init() {
	// DefaultP = plot.New()
	//
	// DefaultP.Title.Text = "roi-yield"
	// DefaultP.X.Label.Text = "roi"
	// DefaultP.Y.Label.Text = "yield"

	if _, err := os.Stat(Path); err != nil {
		err := os.Mkdir(Path, os.ModePerm)
		if err != nil {
			panic(`路径创建失败 Mkdir err: ` + err.Error())
		}
	}
}

func Save(points []interface{}, symbol string, maxX float64) {
	P := plot.New()

	P.Title.Text = "roi-yield"
	P.X.Label.Text = "roi"
	P.Y.Label.Text = "yield"

	_ = plotutil.AddLinePoints(P, points...)
	_ = plotutil.AddLinePoints(P, plotter.XYs{{0, 0}, {maxX * 1.1, 0}})
	for i := 0.1; i < maxX; i += 0.1 {
		_ = plotutil.AddLinePoints(P, plotter.XYs{{i, 0}, {i, 20}})
	}

	if err := P.Save(15*vg.Inch, 6*vg.Inch, path.Join(Path, symbol+time.Now().Format("20060102_150405.png"))); err != nil {
		log.Fatal(err)
	}
}

func randomPoints(n int) plotter.XYs {
	points := make(plotter.XYs, n)
	for i := range points {
		if i == 0 {
			points[i].X = rand.Float64()
		} else {
			points[i].X = points[i-1].X + rand.Float64()
		}
		points[i].Y = points[i].X + 10*rand.Float64()
	}

	return points
}
