package exchange

import (
	"context"
	"fmt"
	"github.com/daihaoxiaofei/balance/config"
	"github.com/daihaoxiaofei/tool"
	"testing"
	"time"
)

// 最好用这个程序提前下载好需要的数据  现场下载总是有些奇怪的问题
func TestGetHistory(t *testing.T) {
	var StartTime = time.Date(2020, 9, 1, 0, 0, 0, 0, time.Local).Unix() * 1e3
	var EndTime = time.Date(2022, 5, 11, 0, 0, 0, 0, time.Local).Unix() * 1e3

	// Coins := []string{`SOL`}
	Coins := []string{`SOL`, `XRP`, `ADA`, `DOGE`, `AVAX`, `DOT`, `BTC`, `ETH`}
	for _, Coin := range Coins {
		fmt.Println(len(GetHistory(Coin+config.Balance.BaseCoin, StartTime, EndTime)))
	}

	// StartTime := time.Date(2021, 5, 1, 0, 0, 0, 0, time.Local).Unix() * 1e3
	// EndTime := time.Date(2021, 5, 1, 20, 0, 0, 0, time.Local).Unix() * 1e3
	// da := GetHistory(`BTCUSDT`, StartTime, EndTime)
	// fmt.Println(len(da))
}

// 币安返回历史1m数据中 开始时间有可能和指定的时间不一样 除上线时间早于指定时间外
// 还有些意外的情况 如指定 StartTime= 1619337600000 但返回数据是从1619340300000开始的 测试了BTCUSDT SOLUSDT均有问题 感觉是币安的坑
func TestGetHistory2(t *testing.T) {
	Kline, err := bian.NewKlinesService().
		Symbol(`BTCUSDT`).
		Interval(`1m`).
		// StartTime(1619337600000-1*3600*1e3).
		StartTime(time.Date(2020, 9, 1, 0, 0, 0, 0, time.Local).Unix() * 1e3).
		Limit(1).
		Do(context.Background())
	if err != nil {
		panic(`b.bc err: ` + err.Error())
	}
	tool.SmartPrint(Kline[0])
}
