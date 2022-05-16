package strategy

import (
	"fmt"
	"github.com/daihaoxiaofei/balance/config"
	"math"
)

// 继承
type BackTest struct {
	Balance
}

func NewBackTest(BCoins []Base) *BackTest {
	// 权重检查
	allWeight := float64(0)
	for _, v := range BCoins {
		allWeight += v.Weight
	}
	if allWeight != 1 {
		panic(fmt.Sprintf("各币种权重之和不为1,现配置为:%f", allWeight))
	}

	t := &BackTest{
		Balance{
			Base:   BCoins,
			Change: make([]Change, len(BCoins), len(BCoins)),
		},
	}

	for k, v := range BCoins {
		t.Change[k].Coin = v.Coin
		t.Coins = append(t.Coins, v.Coin)
		t.Symbols = append(t.Symbols, v.Coin+config.Balance.BaseCoin)
	}

	return t
}

// 根据初始价格分配数量
func (t *BackTest) SetNum(priceMap map[string]float64) {
	// 获取总金额
	allMoney := 10000.0
	// 价格和数量
	for k, v := range t.Change {
		t.Change[k].Price = priceMap[v.Coin]
		t.Change[k].Num = allMoney * t.Base[k].Weight / t.Change[k].Price
		t.Base[k].Num = t.Change[k].Num
	}
}

// 回测 需要传入价格map获取各币种目前价格
func (t *BackTest) BackTest(priceMap map[string]float64) {
	// 价格波动大于一定程度才执行重新平衡
	var reallocate bool
	for _, v := range t.Change {
		if math.Abs(priceMap[v.Coin]-v.Price)/v.Price >= config.Run.Roi {
			reallocate = true
			break
		}
	}

	if reallocate {
		// 获取总金额
		allMoney := 0.0
		for _, v := range t.Change {
			allMoney += priceMap[v.Coin] * v.Num
		}
		for k, v := range t.Change {
			// 刷新价格和数量
			t.Change[k].Num = allMoney * t.Base[k].Weight / priceMap[v.Coin]
			t.Change[k].Price = priceMap[v.Coin]
		}
	}
}

// 结算
func (t *BackTest) CloseAnAccount() [2]float64 {
	// 动态利润率=使用策略:拿着不动
	// 静态利润率=使用策略:时光冻结
	// 动态利润率(以基础货币数量计算)=(现在的总价值-过去的数量放到现在总价值/过去的数量放到现在总价值)
	dynamicBalance := float64(0)
	newBalance := float64(0)
	// 这里的价格取值只能拿最后一次平衡后的价格了 再跑去拿最后`时刻`的价格繁琐好像也没什么差别
	for k, _ := range t.Change {
		// 价格*数量
		dynamicBalance += t.Change[k].Price * t.Base[k].Num // 现在价格*原始数量
		newBalance += t.Change[k].Price * t.Change[k].Num   // 现在价格*现在数量
	}
	return [2]float64{newBalance,dynamicBalance}
}
