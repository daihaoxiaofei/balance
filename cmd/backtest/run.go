package backtest

import (
	"fmt"
	"github.com/daihaoxiaofei/balance/config"
	"github.com/daihaoxiaofei/balance/exchange"
	"github.com/daihaoxiaofei/balance/helpfunc"
	"github.com/daihaoxiaofei/balance/strategy"
	"github.com/spf13/cobra"
	"time"
)

var (
	par     string
	Command = &cobra.Command{
		Use:     "backtest",
		Aliases: []string{"b"}, // 添加备用值
		Short:   "回测",
		Example: "./balance backtest",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

var StartTime = time.Date(2017, 9, 1, 0, 0, 0, 0, time.Local).Unix() * 1e3
var EndTime = time.Date(2022, 5, 15, 0, 0, 0, 0, time.Local).Unix() * 1e3

func init() {
	// 获取命令行配置 默认给了个流水的子命令
	Command.PersistentFlags().StringVarP(&par, "command", "c", "finance", "命令")
}

// go run main.go backtest -command 2
// go run main.go b -c 2
func run() {
	switch par {
	case `1`:
		backtest1()
	case `2`:
		backtest2()
	default:
		fmt.Println(`无效参数`)
	}

}

// 整体币种回测
func backtest1() {
	exclude := []string{`USDT`, `USDC`, `BUSD`, `TUSD`, `USDP`, `PAXG`, `BUSD`, `TUSD`, `USDP`, `PAXG`, `WBTC`, `DAI`,
		`CRO`, `LEO`, `UST`, `APE`, `KCS`, `HT`, `BSV`, `MIOTA`, `GMT`,
		`USDN`, `OKB`, `NEXO`, `XDC`, `KDA`, `GT`, `LDO`, `FEI`, `XYM`} // 排除
	coins := make([]string, 0)
	for _, coin := range helpfunc.GetCoinMsg() {
		if !helpfunc.InArray(coin, exclude) {
			coins = append(coins, coin)
		}
	}
	fmt.Printf("%s\t%s\t%s\t%s\n", "币种", `价格涨幅`, `策略总值`, `不使用策略总值`)

	for _, coin := range coins {
		Init := []strategy.Base{
			{Coin: coin, Weight: 0.5},
			{Coin: `USDT`, Weight: 0.5},
		}
		Ba := strategy.NewBackTest(Init)
		HistoryChan := exchange.GetHistoryChan(Ba.Coins, StartTime, EndTime)
		Ba.SetNum(<-HistoryChan)
		for v := range HistoryChan {
			Ba.BackTest(v)
		}
		re := Ba.CloseAnAccount()
		fmt.Printf("%s\t%f\t%f\t%f\n", coin, Ba.Change[0].Price/(5000/Ba.Base[0].Num), re[0], re[1])
	}
}

// 单币种调参回测  纵向为不同触发平衡价格比  横向为不同币种-稳定币比重
func backtest2() {
	for Roi := 0.1; Roi < 1; Roi += 0.1 {
		config.Run.Roi = Roi
		fmt.Printf("%f\t", Roi)
		for coinRoi := 0.1; coinRoi < 1; coinRoi += 0.1 {
			Init := []strategy.Base{
				{Coin: `XRP`, Weight: coinRoi},
				{Coin: `USDT`, Weight: 1 - coinRoi},
			}
			Ba := strategy.NewBackTest(Init)
			HistoryChan := exchange.GetHistoryChan(Ba.Coins, StartTime, EndTime)
			Ba.SetNum(<-HistoryChan)
			for v := range HistoryChan {
				Ba.BackTest(v)
			}
			re := Ba.CloseAnAccount()

			fmt.Printf("%s\t", Sprint(re[0]))
			// fmt.Printf("%f\t%s\t%f\t%f\t%f\n", Roi, coin, re[0])
		}
		fmt.Println()
	}
}

func Sprint(x float64) string {
	// max := float64( / 7)
	i := int(x*6/150000) + 31
	return fmt.Sprintf("%c[0;0;%dm%f%c[0m", 0x1B, i, x, 0x1B)
}
