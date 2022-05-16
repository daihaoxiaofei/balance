package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Run     run
	Auth    auth
	Proxy   proxy
	Balance balance
	Notify  notify
)

type run struct {
	CoinMarketKey string
	Roi           float64
}

type auth struct {
	ApiKey    string
	SecretKey string
}

type proxy struct {
	Open bool
	Path string
}

type balance struct {
	CoinMsg  []CoinMsg
	BaseCoin string
}
type CoinMsg struct {
	Name   string
	Weight float64
}

type notify struct {
	DD struct {
		Open   bool
		Url    string
		Secret string
	}
	Email struct {
		Open bool
		HOST string
		PORT int
		From string
		Pwd  string
		To   string
	}
}

func init() {
	vp := viper.New()
	if (runtime.GOOS == `linux` && os.Args[0][len(os.Args[0])-5:] == `.test`) ||
		(runtime.GOOS == `windows` && filepath.Base(os.Args[0])[:7] == `___Test`) {
		// fmt.Println(runtime.GOOS, `测试环境配置文件`)
		_, onPath, _, _ := runtime.Caller(0)
		onDir := filepath.Dir(onPath)
		vp.AddConfigPath(onDir)
	} else {
		// fmt.Printf("普通环境: os:%s , Args[0]: %s  , Base: %s \n", runtime.GOOS,os.Args[0],filepath.Base(os.Args[0])[:7])
		// fmt.Println(runtime.GOOS, `os.Args[0]`,os.Args[0])
		vp.AddConfigPath("config")
	}

	vp.SetConfigType("yml")
	vp.SetConfigName("config") // 可以不指定文件名

	err := vp.ReadInConfig()
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}

	err = vp.UnmarshalKey("Run", &Run)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Auth", &Auth)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Balance", &Balance)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Proxy", &Proxy)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Notify", &Notify)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}

}
