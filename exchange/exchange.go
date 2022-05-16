package exchange

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	"github.com/daihaoxiaofei/balance/config"
	"github.com/daihaoxiaofei/balance/helpfunc"
	"github.com/daihaoxiaofei/fcache"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Limit = 720 // 一次请求的历史数据个数 720/24=30天  最高 41天 41*24=984

var (
	bian = NewBiance() // 快速测试使用
	mu   = sync.Mutex{}
)

// 文档: https://binance-docs.github.io/apidocs/spot/cn/#185368440e
func NewBiance() *binance.Client {
	client := binance.NewClient(config.Auth.ApiKey, config.Auth.SecretKey)
	// 代理设置:
	if config.Proxy.Open {
		httpTransport := &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(config.Proxy.Path)
			},
		}
		httpClient := &http.Client{
			Transport: httpTransport,
		}
		client.HTTPClient = httpClient
	}
	return client
}

type History struct {
	Time  int64
	Price float64
}

type historySlice struct {
	data  []History
	index int64
}

// 获取分时k线的切片
func getHistorySlice(fc *fcache.FileCache, symbol string, StartTime int64) historySlice {
	var resI []History
	_ = fc.Remember(`Slice-`+symbol+`-`+time.Unix(StartTime/1e3, 0).Format(`20060102_1504`),
		&resI, func() interface{} {
		RE:
			Kline, err := bian.NewKlinesService().
				Symbol(symbol).
				Interval(`1h`).
				StartTime(StartTime).
				Limit(Limit).
				Do(context.Background())
			if err != nil {
				APIError, ok := err.(*common.APIError)
				if ok {
					if APIError.Code == -1003 {
						fmt.Println(`过于频繁 休眠一分钟后继续请求`)
						time.Sleep(time.Minute) // 休眠一分钟后继续请求
						goto RE
					}
					if APIError.Code == -1121 {
						panic(`币种错误: ` + strings.Replace(symbol, `USDT`, ``, 1))
					}

				}
				fmt.Println(`未知错误 err TypeOf:`, reflect.TypeOf(err))
				panic(`b.bc err: ` + err.Error())
			}

			// 时间准确性检查
			endTime := StartTime + int64(Limit*3600*1e3) // 当前切片的截至时间
			for i := 0; i < len(Kline); i++ {
				if Kline[i].OpenTime > endTime {
					break
				}
				price, _ := strconv.ParseFloat(Kline[i].Open, 64)
				resI = append(resI, History{Kline[i].OpenTime, price})
			}
			return resI
		})
	return historySlice{index: StartTime, data: resI}
}

// 获取分时k线  带缓存  多线程处理
func GetHistory(symbol string, StartTime, EndTime int64) (res []History) {
	_, onPath, _, _ := runtime.Caller(0)
	onDir := filepath.Dir(onPath)
	fc := fcache.NewFC(path.Join(onDir, `cache`))

	fileName := `History-` + symbol + `-` + time.Unix(StartTime/1e3, 0).Format(`20060102_1504`) + `-` +
		time.Unix(EndTime/1e3, 0).Format(`20060102_1504`)

	// 防止多线程调用造成的重复请求 简单判断文件是否存在 若不存在则加锁
	if _, err := os.Stat(path.Join(fc.DirPath, fileName+fc.Suffix)); err != nil {
		mu.Lock()
		defer mu.Unlock()
	}

	err := fc.Remember(fileName, &res, func() interface{} {
		fmt.Println(symbol + ` 历史数据采集中...`)

		historySlices := make([]historySlice, 0)
		wg := sync.WaitGroup{}
		// i:开始时间; Limit:个数;  3600:一小时秒数;  1e3: 毫秒
		for i := StartTime; i < EndTime; i += Limit * 3600 * 1e3 {
			wg.Add(1)
			go func(i int64) {
				defer wg.Done()
				resI := getHistorySlice(fc, symbol, i)
				historySlices = append(historySlices, resI)
			}(i)
		}
		wg.Wait() // 等待所有请求处理完成
		// 排序
		for i := 0; i < len(historySlices)-1; i++ {
			for j := 0; j < len(historySlices)-1-i; j++ {
				if historySlices[j].index > historySlices[j+1].index {
					historySlices[j], historySlices[j+1] = historySlices[j+1], historySlices[j]
				}
			}
		}
		// 按顺序取数据
		for _, v := range historySlices {
			res = append(res, v.data...)
		}
		if len(res) == 0 {
			panic(symbol + ` 没有数据`)
		}
		fmt.Println(symbol + ` 历史数据采集完成`)
		return res
	})
	if err != nil {
		panic(`fc.Remember err: ` + err.Error())
	}
	return
}

// 获取分时k线 channel
func GetHistoryChan(coins []string, StartTime, EndTime int64) <-chan map[string]float64 {
	ch := make(chan map[string]float64, 1)
	go func() {
		priceMap := make(map[string][]History)
		for _, coin := range coins {
			if coin != `USDT` {
				priceMap[coin] = GetHistory(coin+config.Balance.BaseCoin, StartTime, EndTime)
			}
		}
		// 协调各各币种的时间 只初步判断一下数量
		historyLen := len(priceMap[coins[0]])
		for i := 1; i < len(coins); i++ {
			if coins[i] != `USDT` && historyLen != len(priceMap[coins[i]]) {
				panic(`选取的币种历史记录数量不同`)
			}
		}
		for i := 0; i < historyLen; i++ {
			mapHist := make(map[string]float64)
			for _, coin := range coins {
				if coin == `USDT` {
					mapHist[coin] = 1
				} else {
					mapHist[coin] = priceMap[coin][i].Price
				}
			}
			ch <- mapHist
		}
		close(ch)
	}()

	return ch
}

// 下单
func CreateOrder(coin string, Quantity, price float64) {
	// 判断方向
	var sideType binance.SideType
	if Quantity > 0 {
		sideType = binance.SideTypeBuy
	} else { // 卖出
		sideType = binance.SideTypeSell
	}
	// 过滤器
	err := FiltersMarket(coin+config.Balance.BaseCoin, &Quantity, price)
	if err != nil {
		fmt.Println(`过滤器未通过`, err)
		return
	}
	res, err := bian.NewCreateOrderService().
		Symbol(coin + `USDT`).
		Side(sideType).
		Quantity(fmt.Sprintf("%f", Quantity)).
		Type(binance.OrderTypeMarket).
		Do(context.Background())
	if err != nil {
		if APIError, ok := err.(*common.APIError); ok {
			if APIError.Code == -1100 {
				fmt.Println(Quantity)
				fmt.Println(fmt.Sprintf("%0.2f", Quantity))
			}
		}
		panic(`NewCreateOrderService err:` + err.Error())
	}
	fmt.Println(`下单成功`, res)
}

// 对指定交易对的数据进行过滤: 加工和阻拦 市价单情况下只对数量简单检查和加工  价格作为辅助参数
func FiltersMarket(symbols string, quantity *float64, price float64) error {
	Filters, err := GetFilters([]string{`LUNABUSD`})
	if err != nil {
		return err
	}
	for k, v := range Filters { // Filters:map[string][]map[string]interface{}
		if k == symbols { // 是指定币种
			// fmt.Println(k)
			for _, val := range v { // v:[]map[string]interface{}
				// 每组过滤器
				switch val[`filterType`] {
				// case `PRICE_FILTER`: // 价格过滤器 用于检测订单中 price 参数的合法性。包含以下三个部分:
				// 	minPriceStr, _ := val[`minPrice`].(string)
				// 	maxPriceStr, _ := val[`maxPrice`].(string)
				// 	tickSizeStr, _ := val[`tickSize`].(string)
				//
				// 	minPrice, _ := strconv.ParseFloat(minPriceStr, 10)
				// 	maxPrice, _ := strconv.ParseFloat(maxPriceStr, 10)
				// 	tickSize, _ := strconv.ParseFloat(tickSizeStr, 10)
				// 	if *price < minPrice {
				// 		return errors.New(fmt.Sprintf(`价格:过低 , %f, %f`, *price, minPrice))
				// 	}
				// 	if *price > maxPrice {
				// 		return errors.New(fmt.Sprintf(`价格:过高 , %f, %f`, *price, maxPrice))
				// 	}
				// 	*price = float64(int(*price/tickSize)) * tickSize // 调整步长
				// case `LOT_SIZE`: // 订单尺寸 对订单中的 quantity 也就是数量参数进行合法性检查
				// 	minQtyStr, _ := val[`minQty`].(string)
				// 	maxQtyStr, _ := val[`maxQty`].(string)
				// 	stepSizeStr, _ := val[`stepSize`].(string)
				//
				// 	minQty, _ := strconv.ParseFloat(minQtyStr, 10)
				// 	maxQty, _ := strconv.ParseFloat(maxQtyStr, 10)
				// 	stepSize, _ := strconv.ParseFloat(stepSizeStr, 10)
				// 	if *quantity < minQty {
				// 		return errors.New(fmt.Sprintf(`价格:过低 , %f, %f`, *quantity, minQty))
				// 	}
				// 	if *quantity > maxQty {
				// 		return errors.New(fmt.Sprintf(`价格:过高 , %f, %f`, *quantity, maxQty))
				// 	}
				// 	*quantity = float64(int(*quantity/stepSize)) * stepSize // 调整步长
				case `MARKET_LOT_SIZE`: // 订单尺寸 MARKET订单定义了数量(即拍卖中的"手数")规则
					minQtyStr, _ := val[`minQty`].(string)
					maxQtyStr, _ := val[`maxQty`].(string)
					stepSizeStr, _ := val[`stepSize`].(string)

					minQty, _ := strconv.ParseFloat(minQtyStr, 10)
					maxQty, _ := strconv.ParseFloat(maxQtyStr, 10)
					stepSize, _ := strconv.ParseFloat(stepSizeStr, 10)
					if *quantity < minQty {
						return errors.New(fmt.Sprintf(`价格:过低 , %f, %f`, *quantity, minQty))
					}
					if *quantity > maxQty {
						return errors.New(fmt.Sprintf(`价格:过高 , %f, %f`, *quantity, maxQty))
					}
					*quantity = float64(int(*quantity/stepSize)) * stepSize // 调整步长
				case `MIN_NOTIONAL`: // 最小名义价值(成交额) 暂时挂的是市价单 只能预估不够精准
					if val[`applyToMarket`] == true {
						minNotionalStr, _ := val[`minNotional`].(string)
						minNotional, _ := strconv.ParseFloat(minNotionalStr, 10)
						if *quantity*price < minNotional {
							return errors.New(fmt.Sprintf(`成交额:过低 , %f,minNotional: %f`, *quantity*price, minNotional))
						}
					}
				default:
					// fmt.Println(`其他过滤器?:`, val)
				}
			}
		}
	}
	// filterType PRICE_FILTER
	// minPrice 0.00001000
	// maxPrice 10000.00000000
	// tickSize 0.00001000
	//
	// filterType PERCENT_PRICE
	// multiplierUp 5
	// multiplierDown 0.2
	// avgPriceMins 5
	//
	// filterType LOT_SIZE
	// minQty 0.01000000
	// maxQty 9000000.00000000
	// stepSize 0.01000000
	//
	// filterType MIN_NOTIONAL
	// minNotional 10.00000000
	// applyToMarket true
	// avgPriceMins 5
	//
	// filterType ICEBERG_PARTS
	// limit 10
	//
	// filterType MARKET_LOT_SIZE
	// minQty 0.00000000
	// maxQty 3366034.49077083
	// stepSize 0.00000000
	//
	// filterType TRAILING_DELTA
	// minTrailingAboveDelta 10
	// maxTrailingAboveDelta 2000
	// minTrailingBelowDelta 10
	// maxTrailingBelowDelta 2000
	//
	// maxNumOrders 200
	// filterType MAX_NUM_ORDERS
	//
	// filterType MAX_NUM_ALGO_ORDERS
	// maxNumAlgoOrders 5

	return nil
}

// 获取账户余额
func GetCoinNum(symbols []string) (map[string]float64, error) {
	Account, err := bian.NewGetAccountService().Do(context.Background())
	if err != nil { //
		return nil, err
	}
	res := make(map[string]float64)
	for _, each := range Account.Balances {
		if helpfunc.InArray(each.Asset, symbols) {
			price, _ := strconv.ParseFloat(each.Free, 64)
			res[each.Asset] = price
		}
	}
	return res, nil
}

// 获取币种现价
func GetCoinPrice(symbols []string) (map[string]float64, error) {
	SymbolPrice, err := bian.NewListPricesService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	res := make(map[string]float64)
	for _, each := range SymbolPrice {
		if helpfunc.InArray(each.Symbol, symbols) {
			price, _ := strconv.ParseFloat(each.Price, 64)
			res[strings.Replace(each.Symbol, `USDT`, ``, 1)] = price // 截取掉 usdt
		}
	}
	if helpfunc.InArray(`USDTUSDT`, symbols) {
		res[`USDT`] = 1
	}
	// 检查数据完整性
	for _, symbol := range symbols {
		if _, ok := res[strings.Replace(symbol, `USDT`, ``, 1)]; !ok {
			panic(`远程获取的币价数据中缺少:` + symbol)
		}
	}
	return res, nil
}

// GetFilters 过滤器 参考:https://binance-docs.github.io/apidocs/spot/cn/#api-3
func GetFilters(symbols []string) (map[string][]map[string]interface{}, error) {
	symbolInfos := make(map[string][]map[string]interface{}, len(symbols))
	info, err := bian.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		// if urlErr,ok:=err.(*url.Error);ok{
		// 	tool.SmartPrint( urlErr)
		// }
		// fmt.Println(reflect.TypeOf(err))
		return nil, err
	}
	for _, each := range info.Symbols {
		if helpfunc.InArray(each.Symbol, symbols) {
			symbolInfos[each.Symbol] = each.Filters
		}
	}
	return symbolInfos, nil
}
