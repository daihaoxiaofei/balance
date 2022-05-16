package strategy

import (
	"fmt"
	"github.com/daihaoxiaofei/balance/config"
	"github.com/daihaoxiaofei/balance/exchange"
	"math"
)

// 平衡策略 扩展为多币种平衡  需要加入USDT做价格指标
// 固定币种间的比例 价格变化一定百分比时 通过买卖将比例还原
// 例: BNB-USDT 位 50%:50% 目前价格350 则持有100个BNB+35000USDT
// 当价格变成400时 BNB应有数量=(35000+40000)/2/400=93.75 此时需要卖掉100-93.75=6.25
// 当价格变成300时 BNB应有数量=(35000+30000)/2/300=108.33 此时需要买入 8.33
// 如此反复
// 可控变量: 触发平衡的价格波动比例  需要维持的币种见比例

type Change struct {
	Coin  string  // 币种
	Price float64 // 最近一次调整时的价格 做对比用的
	Num   float64 // 数量
}

// 初始数据
type Base struct {
	Coin   string
	Num    float64
	Weight float64 // 权重
}

type Balance struct {
	Change  []Change
	Base    []Base                              // 记录初始状态
	Coins   []string                            // 记录初始状态
	Symbols []string                            // 记录初始状态
	Filters map[string][]map[string]interface{} // 过滤器
}

func NewBalance(BCoins []Base) *Balance {
	// 权重检查
	allWeight := float64(0)
	for _, v := range BCoins {
		allWeight += v.Weight
	}
	if allWeight != 1 {
		panic(fmt.Sprintf("各币种权重之和不为1,现配置为:%f", allWeight))
	}

	t := &Balance{
		Base:   BCoins,
		Change: make([]Change, len(BCoins), len(BCoins)),
	}

	for k, v := range BCoins {
		t.Change[k].Coin = v.Coin
		t.Coins = append(t.Coins, v.Coin)
		t.Symbols = append(t.Symbols, v.Coin+config.Balance.BaseCoin)
	}
	// 获取过滤器限制
	var err error
	t.Filters, err = exchange.GetFilters(t.Symbols)
	if err != nil {
		panic(`初始化获取过滤器失败: ` + err.Error())
	}
	// 设置一个初始离谱的价格
	for k := range BCoins {
		t.Change[k].Price = math.MaxFloat64 / 1e8
	}

	return t
}

// 执行策略
func (t *Balance) Do() {
	// 获取各币种目前价格
	priceMap, err := exchange.GetCoinPrice(t.Symbols)
	if err != nil {
		fmt.Println(`getCoinPrice: err: `, err)
		return
	}
	// 价格波动大于一定程度才执行重新平衡
	var reallocate bool
	for _, v := range t.Change {
		if math.Abs(priceMap[v.Coin]-v.Price)/v.Price >= config.Run.Roi {
			reallocate = true
			break
		}
	}

	if reallocate {
		// 获取各币种目前数量
		NumMap, err := exchange.GetCoinNum(t.Coins)
		if err != nil {
			fmt.Println(`getCoinNum: err: `, err)
			return
		}
		// 获取总金额
		allMoney := 0.0
		for k, v := range NumMap {
			allMoney += priceMap[k] * v
		}
		for k, v := range t.Change {
			// 刷新价格和数量
			t.Change[k].Num = NumMap[v.Coin]
			t.Change[k].Price = priceMap[v.Coin]
			// 差额= 新-旧 = 总金额*权重/价格 - 旧数量
			diffNum := allMoney*t.Base[k].Weight/priceMap[v.Coin] - NumMap[v.Coin]
			// fmt.Println(`差额`, v.Coin, diffNum)
			if v.Coin != `USDT` {
				// t.ChangeNum(v.Coin, diffNum, priceMap[v.Coin])
				exchange.CreateOrder(v.Coin, diffNum, priceMap[v.Coin])
			}
		}
	}
}
