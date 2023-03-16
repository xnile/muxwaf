# MuxWaf

MuxWaf是基于OpenResty实现的一款防CC的高性能WAF。

## MuxWAF能做什么？

![](https://raw.githubusercontent.com/xnile/muxwaf/bd0e1f8578c02d1b78e13ba704cdd62ef50dbc50/screenshot/screen01.png)

* 频率限制防CC
* 白名单功能
* IP及地域级IP黑名单功能
* 支持API管理
* 自带WEB管理后台

## 组件
* guard:  waf引擎，基于openresty开发。
* apiserver: 管理后台api，基于golang开发。
* ui: 管理后台前端页面，基于antdv开发。

## 安装
### 快速体验

需要docker和docker-compose环境。

* git clone https://github.com/xnile/muxwaf ./
* cd muxwaf
* make run
* waf 入口：http://localhost:8080/
* 管理后台地址：http://localhost:8000/ ，默认用户名和密码：admin/admin@123
* 登录管理后台切到`系统管理->节点管理->添加节点`，输入 `guard/8083`添加waf节点
* 添加网站，后台位置：`网站管理->网站管理->新增网站`。
* 通过绑Host的方式将上边网站域名指向本机，然后访问http://域名:8080 就可以开始体验了。