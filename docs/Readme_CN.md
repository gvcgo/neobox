[En](https://github.com/moqsien/neobox)

---------------------------

## Neobox是什么？

Neobox是一个强大的免费梯子管理器。可以提供非常棒的免费梯子方案。
它支持几乎所有常见协议，例如，vmess/vless/trojan/shadowsocks/shadowsocksr/wireguard.

## Neobox能做什么？

- 检测免费VPN的可用性，不是ping，而是直接请求，所以筛选精准度高.
- 筛选Cloudflare的CDN节点IP，可以自动将结果用于加速Wireguard/Warp/Warp+/EdgeTunnel.
- [Cloudflare EdgeTunnel](https://github.com/3Kmfi6HP/EDtunnel)单独维护，还可以手动添加其他稳定的梯子(比如自建的VPS等).
- Cloudflare 的wireguard注册，可以升级到Warp，Warp+，请自行检索Warp+ license key.
- 一个交互式Shell，提供了以上所有功能的命令.

## Shell命令展示

```bash
$neobox shell
>>> help

Commands:
  add            Add proxies to neobox mannually. # 添加自己的固定代理 
  added          Add edgetunnel proxies to neobox. # 添加EdgeTunnel的vless代理
  cfip           Test speed for cloudflare IPv4s. # 筛选cloudflare CDN节点
  clear          clear the screen # 清屏
  exit           exit the program # 退出shell
  filter         Start filtering proxies by verifier manually. # 手动开启免费IP筛选
  gc             Start GC manually. # 手动触发GC，降低内存占用
  geoinfo        Install/Update geoip&geosite for neobox client. # 手动下载/更新geoip和geosite信息
  graw           Manually dowload rawUri list(conf.txt from gitlab) for neobox client. # 手动触发原始的免费代理列表下载
  guuid          Generate UUIDs. # 生成uuid
  help           display help # 显示帮助信息
  restart        Restart the running neobox client with a chosen proxy. [restart proxy_index] # 使用指定序号的代理重启
  rmproxy        Remove a manually added proxy [manually or edgetunnel]. # 删除指定的手动添加IP，格式rmproxy address:port
  setkey         Setup rawlist encrytion key for neobox. [With no args will set key to default value] # 必须！！！设置key，用于解密原始列表
  setping        Setup ping without root for Linux. ## Linux系统下，免root权限的ping设置
  show           Show neobox info. # 展示可用代理列表和统计信息，neobox当前运行状态等等
  start          Start a neobox client/keeper. # 开启neobox
  stop           Stop neobox client. # 停止neobox
  wireguard      Register wireguard account and update licenseKey to Warp+ [if a licenseKey is specified]. # 注册wireguard，如果指定了warp+的license key，则升级到warp+账户，配合cloudflare节点筛选，可以加速github、google等的访问


>>> added help # 显示单个命令的帮助信息

Add edgetunnel proxies to neobox.
 options:
  --uuid=xxx; alias:-{u}; description: uuid for edge tunnel vless.
  --address=xxx; alias:-{a}; description: domain/ip for edge tunnel.
 args:
  full raw_uri[vless://xxx@xxx?xxx]
```

## 在Shell中查看Neobox的运行状态

```bash
>>> show

RawList[5344@2023-09-10 10:06:44] vmess[2646] vless[317] trojan[660] ss[1638] ssr[83]
Pinged[587@2023-09-10 12:38:53] vmess[348] vless[53] trojan[86] ss[96] ssr[4]
Final[13@2023-09-10 12:41:13] vmess[4] vless[7] trojan[1] ss[1] ssr[0]
Database: History[27] EdgeTunnel[1] Manually[0]

idx  pxy                                       loc rtt  src
0  vmess://19.kccic2pa.xyz:50019              CHN  1082  verified
1  vmess://cfcdn2.sanfencdn.net:2052          USA  2336  verified
2  vmess://172.67.135.195:443                 USA  1642  verified
3  vmess://103.160.204.80:8080                APA  2435  verified
4  vless://jp.caball.link:443                 USA  1445  verified
5  vless://104.23.139.0:80                    USA  2622  verified
6  vless://172.66.47.84:443                   USA  965  verified
7  vless://fiberlike.aurorainiceland.com:443  USA  1869  verified
8  vless://172.67.187.111:443                 USA  1756  verified
9  vless://172.66.44.172:2096                 USA  2185  verified
10  vless://729.outline-vpn.cloud:443          USA  2635  verified
11  trojan://108.181.23.249:443                CAN  829  verified
12  ss://54.169.127.65:443                     SGP  439  verified
w0  wireguard://108.162.193.239:2096          USA 157  wireguard
e0  vless://xxx.xxx.xx:8443                   USA  255  edtunnel

NeoBox[running @vless://172.64.151.81:443 (Mem: 4098MiB)] Verifier[completed] Keeper[running]
LogFileDir: C:\Users\moqsien\.neobox\log_files
```

## 如何安装？

```bash
go install -tags "with_wireguard with_shadowsocksr with_utls with_gvisor with_grpc with_ech with_dhcp" github.com/moqsien/neobox/example/neobox@latest
```

或者安装[gvc](https://github.com/moqsien/gvc)

## 注意事项
想要使用免费代理，则必须获取解密Key. 当前用于解密的Key是 **5lR3hcN8Zzpo1nzI**.

设置解密Key：
```bash
>>>setkey 5lR3hcN8Zzpo1nzI

```

默认本地Inbound:
```text
http://localhost:2023
```

## 特别声明
**本项目仅作学习研究使用，无任何谋利和不良引导行为。更无意违反任何国家的法律法规。**
**请大陆使用者谨慎使用，切勿向大陆内网中分享一些有损国家利益的内容！**
**学习强国，有你有我！**

## 感谢
- [xray-core](https://github.com/XTLS/Xray-core)
- [sing-box](https://github.com/SagerNet/sing-box)
- [wgcf](https://github.com/ViRb3/wgcf)
- [CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest)
- [EDtunnel](https://github.com/3Kmfi6HP/EDtunnel)
- [vpnparser](https://github.com/moqsien/vpnparser)
- [gscraper](https://github.com/moqsien/gscraper)
- [goktrl](https://github.com/moqsien/goktrl)
