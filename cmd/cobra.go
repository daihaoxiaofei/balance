package cmd

import (
	"errors"
	"fmt"
	"github.com/daihaoxiaofei/balance/cmd/backtest"
	"github.com/daihaoxiaofei/balance/cmd/run"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "today",
	Short: "初始说明",
	Args: func(cmd *cobra.Command, args []string) error { // 检查参数
		if len(args) < 1 {
			return errors.New("需要一个参数")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", "开始项目...")
	},
}

func init() {
	// 加载子命令
	rootCmd.AddCommand(backtest.Command)
	rootCmd.AddCommand(run.Command)

}

// apply commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
