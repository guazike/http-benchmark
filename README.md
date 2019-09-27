# http-benchmark
benchmark for testing http request.

## 简介
  一个http压力测试工具。只需要在配置文件里填充一些测试路径和参数等信息，就可以直接测试了。<br>
具体参照<a href="https://github.com/guazike/http-benchmark/blob/master/src/configs/config_http_demo.json">配置样本</a>。<br>
可以设置测试账号范围、密码<br>
可以指定每个请求的body header coockie<br>
可以设置登录，然后再按顺序执行一系列请求，然后再随机执行一些列请求<br>
可以为每个请求可以设置执行次数或无限次，请求是否启用<br>
将为每一个测试用户建立一个会话，每个会话独立执行配置中设置的各种request<br>

## win64系统可直接使用
修改bin/configs/config_http_demo.json
运行bin/http-benchmark.exe

其它系统可按下面说明编译运行：

## 编译
依赖：go sdk
cd src<br>
go build http-benchmark.go

## 运行
先配置configs/config_http_demo.json，然后运行即可

# win
直接运行：http-benchmark.exe，将使用默认配置configs/config_http_demo.json<br>
也可以用命令行运行: http-benchmark someTest.json ，将使用configs里的someTest.json配置

# linux
运行：./http-benchmark，将使用默认配置configs/config_http_demo.json<br>
也可以指定配置运行: http-benchmark someTest.json ，将使用configs里的someTest.json配置
