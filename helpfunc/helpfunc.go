package helpfunc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daihaoxiaofei/balance/config"
	"github.com/daihaoxiaofei/fcache"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func InArray(symbol string, symbols []string) bool {
	for _, s := range symbols {
		if s == symbol {
			return true
		}
	}
	return false
}

// 用于失败时重新尝试运行 指定重试次数和休眠时间
func Retry(tryNumber int, sleep float64, callback func() error) error {
	var res = errors.New("重试未知错误")

	for i := 1; i <= tryNumber; i++ {
		err := callback()
		if err == nil {
			return nil
		}
		// 只有网络异常才重试
		_, ok := err.(net.Error)
		if !ok {
			res = err
			break
		}
		if i == tryNumber {
			return errors.New(fmt.Sprintf("重试超过次数: %d", tryNumber))
		}
		time.Sleep(time.Duration(sleep))
	}

	return res
}


// 获取币种信息 主要要个币种排名
func GetCoinMsg() []string {
	var out []byte
	_ = fcache.Remember(`coinmarketcap`, &out, func() interface{} {
		// 文档 https://coinmarketcap.com/api/documentation/v1/#tag/cryptocurrency
		UrlStr := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"
		// 代理设置:
		var client *http.Client
		if config.Proxy.Open {
			httpTransport := &http.Transport{
				Proxy: func(_ *http.Request) (*url.URL, error) {
					return url.Parse(config.Proxy.Path)
				},
			}
			client = &http.Client{
				Transport: httpTransport,
			}
		} else {
			client = &http.Client{}
		}

		q := url.Values{}
		q.Add("start", "1")
		q.Add("limit", "100")
		// q.Add("convert", "CNY")

		// 提交请求
		request, err := http.NewRequest("GET", UrlStr, nil)
		if err != nil {
			panic(`http.NewRequest err` + err.Error())
		}
		// 增加header选项
		request.Header.Set("Accepts", "application/json")
		request.Header.Add("X-CMC_PRO_API_KEY", config.Run.CoinMarketKey)
		request.URL.RawQuery = q.Encode()

		// 处理返回结果
		resp, err := client.Do(request)
		if err != nil {
			panic("Error sending request to server")
		}
		defer resp.Body.Close()

		// 读取返回值
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(`ioutil.ReadAll err` + err.Error())
		}
		return body
	})
	// 转结构体
	res := struct {
		Data []struct {
			// Id     int  // id不能代表排名
			Symbol string
			// Quote  struct {
			// 	USD struct {
			// 		Price float64
			// 	}
			// 	CNY struct {
			// 		Price float64
			// 	}
			// }
		}
	}{}
	if err := json.Unmarshal(out, &res); err != nil {
		panic(`Unmarshal err` + err.Error())
	}
	coins := make([]string, 0, len(res.Data))
	for _, v := range res.Data {
		coins = append(coins, v.Symbol)
		// fmt.Println(v.Symbol, v.Quote.CNY.Price)
	}
	return coins
}

// 获取币种历史数据
func getCoinHistory(Unix int64) {
	var out []byte
	_ = fcache.Remember(`CoinHistory-`+time.Unix(Unix, 0).Format(`20060102`), &out, func() interface{} {
		// 文档 https://coinmarketcap.com/api/documentation/v1/#tag/cryptocurrency
		UrlStr := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/historical"

		httpTransport := &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(`socks5://192.168.3.85:10808`)
			},
		}
		client := &http.Client{
			Transport: httpTransport,
		}

		q := url.Values{}
		q.Add("data", strconv.FormatInt(Unix*1e3, 10))
		q.Add("start", "1")
		q.Add("limit", "1")
		// q.Add("convert", "BTC")
		q.Add("convert", "USD")

		// 提交请求
		request, err := http.NewRequest("GET", UrlStr, nil)
		if err != nil {
			panic(`http.NewRequest err` + err.Error())
		}
		// 增加header选项
		request.Header.Set("Accepts", "application/json")
		request.Header.Add("X-CMC_PRO_API_KEY", config.Run.CoinMarketKey)
		request.URL.RawQuery = q.Encode()

		// 处理返回结果
		resp, err := client.Do(request)
		if err != nil {
			panic("Error sending request to server")
		}
		defer resp.Body.Close()

		// 读取返回值
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(`ioutil.ReadAll err` + err.Error())
		}
		return body
	})
	fmt.Println(string(out))
	// 转结构体
	res := struct {
		Data []struct {
			// Id     int  // id不能代表排名
			Symbol string
			Quote  struct {
				USD struct {
					Price float64
				}
				BTC struct {
					Price float64
				}
			}
		}
	}{}
	if err := json.Unmarshal(out, &res); err != nil {
		panic(`Unmarshal err` + err.Error())
	}

	for _, v := range res.Data {
		fmt.Println(v.Symbol, v.Quote.USD.Price)
	}
}
