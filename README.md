# socket5-go
1.学习go语言和socket编程
2.自己编写一个基于socket5的反向代理协议
3.支持SNI擦拭和篡改


使用方法：
1.客户端(运行在带公网的服务器上，比如软路由，自家的PC)
./xsocks5 -L -R 

2.服务器端（运行在不带公网的VPS上，比如google的cloude shell）
./xsocks5 ${address} -----address 是你的带公网的服务器地址，可以是软路由，PC。
