# balance

基于币安交易所实现的动态平衡策略:

    固定币种间的比例 价格变化一定百分比时 通过买卖将比例还原
    例: BNB-USDT 为 50%:50% 目前价格350 则持有100个BNB+35000USDT
    当价格变成400时 BNB应有数量=(35000+40000)/2/400=93.75 此时需要卖掉100-93.75=6.25
    当价格变成300时 BNB应有数量=(35000+30000)/2/300=108.33 此时需要买入 8.33
    如此反复

## 使用

    1 配置config/config.yaml
    2 运行 go run mian.go
    3 没问题的话打包放服务器或本地跑就可以了
    
## 说明

    凡是设置中涉及到的币种都将按照权重重新平衡 意味着您无法保留一定的数量做备用,
    可转移至子账户或资金账户以隔离资金

    
## 待完善

    1 请求出现网络问题需要添加适当的容错
    2 区分何时程序需要停止 其他意外情况下可记录并稳定运行
    3 告警和通知覆盖到位

## 回测情况可查阅 doc/回测报告.md  

    总的来说不看好 对市场行情要求较高 大体横盘才可以
    否则收益不能得到保证(不如拿着不动) 看以后能不能有哪方面优化吧

    