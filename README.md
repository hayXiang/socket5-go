# xsocket5-go
## 目的
1.学习go语言和socket编程    
2.自己编写一个基于socket5的反向代理协议   
3.支持TLS协议的SNI擦拭和篡改  

## 使用方法：
#### 1.服务端 (运行在带公网的机器上，比如软路由，自家的PC)
./xsocks5 -L -R -A ":5201" -S ":8888"

#### 2.客户端  （运行在不带公网的VPS上，比如google的cloude shell）
./xsocks5 ${address} -----address 是你的带公网的服务器地址，可以是软路由，PC。比如www.xsocks5.com:5201   

## 案例：
### goorm
1. goorm 的容器运行， ./xsocks5 www.xsocks5.com:5201 
2. 家中路由器运行， ./xsocks5 -L -R -A ":5201" -S ":8888" （默认开启了9999->22的端口映射 ）   
3. 连通后，就可以在路由器上操作
   3.1 curl --socks5 127.0.0.1:8888 https://www.xxxx.com 
   3.2 ssh 127.0.0.1 -p 9999 
