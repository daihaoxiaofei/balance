package run

import (
	"fmt"
	"github.com/daihaoxiaofei/balance/config"
	"github.com/daihaoxiaofei/balance/strategy"
	"github.com/spf13/cobra"
	"time"
)

var (
	Command = &cobra.Command{
		Use:     "run",
		Aliases: []string{``, "r"}, // 添加备用值
		Short:   "运行主程序",
		Example: "./balance run",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

// go run main.go run
func run() {
	Init := make([]strategy.Base, 0, len(config.Balance.CoinMsg))
	for _, coin := range config.Balance.CoinMsg {
		Init = append(Init, strategy.Base{Coin: coin.Name, Weight: coin.Weight})
	}
	Ba := strategy.NewBalance(Init)
	fmt.Println(Ba)
	// 时钟
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		Ba.Do()
		<-ticker.C
	}
}
