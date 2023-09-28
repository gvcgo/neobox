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
  dedge          Download rawList for a specified edgeTunnel proxy [dedge proxy_index]. # 为指定的EdgeTunnel下载rawList
  domain         Download selected domains file for edgeTunnels. # 下载edgeTunnels的domains列表
  exit           exit the program # 退出shell
  filter         Start filtering proxies by verifier manually. # 手动开启免费IP筛选
  gc             Start GC manually. # 手动触发GC，降低内存占用
  geoinfo        Install/Update geoip&geosite for neobox client. # 手动下载/更新geoip和geosite信息
  graw           Manually dowload rawUri list(conf.txt from gitlab) for neobox client. # 手动触发原始的免费代理列表下载
  guuid          Generate UUIDs. # 生成uuid
  help           display help # 显示帮助信息
  parse          Parse rawUri of a proxy to xray-core/sing-box outbound string [xray-core by default]. # 将某个rawUri解析成xray-core或者sing-box的outbound字符串并显示
  pingd          Ping selected domains for edgeTunnels. # 优选edgeTunnels的domain列表
  qcode          Generate QRCode for a chosen proxy. [qcode proxy_index] # 为指定序号的代理生成二维码，方便手机端(例如, NekoBox等)进行扫码
  remove         Remove a manually added proxy [manually or edgetunnel]. # 删除指定的手动添加IP，格式rmproxy address:port
  restart        Restart the running neobox client with a chosen proxy. [restart proxy_index] # 使用指定序号的代理重启
  setkey         Setup rawlist encrytion key for neobox. [With no args will set key to default value] # 必须！！！设置key，用于解密原始列表
  setping        Setup ping without root for Linux. ## Linux系统下，免root权限的ping设置
  show           Show neobox info. # 展示可用代理列表和统计信息，neobox当前运行状态等等
  start          Start a neobox client/keeper. # 开启neobox
  stop           Stop neobox client. # 停止neobox
  sys-proxy      To enable or disable System Proxy. # 开启/停止  以neobox作为系统代理(即全局代理)
  wireguard      Register wireguard account and update licenseKey to Warp+ [if a licenseKey is specified]. # 注册wireguard，如果指定了warp+的license key，则升级到warp+账户，配合cloudflare节点筛选，可以加速github、google等的访问


>>> added help # 显示单个命令的帮助信息

Add edgetunnel proxies to neobox.
 options:
  --uuid=xxx; alias:-{u}; description: uuid for edge tunnel vless.
  --address=xxx; alias:-{a}; description: domain/ip for edge tunnel.
 args:
  full raw_uri[vless://xxx@xxx?xxx]


>>> restart help # 选择一个代理重启neobox

Restart the running neobox client with a chosen proxy. [restart proxy_index]
 options:
  -showchosen; alias:-{p,proxy}; description: show the chosen proxy or not. # 是否显示选择的proxyItem
  -showconfig; alias:-{c,config}; description: show config in result or not. # 是否显示最终的config字符串
  -usedomains; alias:-{d,domain}; description: use selected domains for edgetunnels. # 使用edgeTunnel重启时使用优选domain，否则使用优选IP
  -usesingbox; alias:-{s,sbox}; description: force to use sing-box as client. # 是否强制使用xray-core作为客户端
 args:
  choose a specified proxy by index.
```

**注意**：

- **二维码**: 你可以用qcode子命令来生成某个代理的二维码。然后在手机端使用诸如NekoBox之类的App扫描二维码，就可以分享neobox的代理筛选结果了。

- **Wireguard**: 你可以通过"wireguard"子命令来自动注册免费的wireguard账户，免费的wireguard账户有流量限制. 如果使用"wireguard xxx"来指定一个Warp+注册码, 
你注册的wireguard免费账号就可以升级到Warp+账号，Warp+账号总流量24.6PB，约等于不限流量. Warp+账号可以从Telegram的[@generatewarpplusbot](https://t.me/generatewarpplusbot)机器人获取.
Warp+注册码长这样："Key: 8YK01D7V-Ij6z50g7-eZ5l2R36 (24598562 GB)".

- **Cloudflare EdgeTunnel**: 使用Cloudflare的[pages/workers](https://dash.cloudflare.com/login)部署[EdgeTunnel](https://github.com/3Kmfi6HP/EDtunnel).具体自行使用搜索引擎检索. 

- **Cloudflare IP/Domain**: Cloudflare相关的IP/Domain优选，支持Warp+和EdgeTunnel的IP/Domain优选，实现网络加速.

- **一些强大的免费资源**: [cloudflare](https://github.com/moqsien/neobox/blob/main/docs/cloudflare.md)

## 在Shell中查看Neobox的运行状态

```bash

>>> show
RawList[3804@2023-09-25 12:58:59] vmess[2163] vless[544] trojan[368] ss[682] ssr[47]
Pinged[359@2023-09-26 13:27:04] vmess[192] vless[40] trojan[51] ss[75] ssr[1]
Final[13@2023-09-26 13:28:28] vmess[1] vless[12] trojan[0] ss[0] ssr[0]
Database: History[341] EdgeTunnel[2] Manually[0]

========================================================================
NeoBox[running @vless://www.wongnai.com:443 (Mem: 4991MiB)] Verifier[running] Keeper[running]
LogFileDir: C:\Users\xxxxxx\.gvc\proxy_files\neobox_logs

========================================================================
Press 'Up/k · Down/j' to move up/down or 'q' to quit.
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Index   Proxy                                                         Location  RTT     Source           │
│──────────────────────────────────────────────────────────────────────────────────────────────────────────│
│ 0      vmess://mf01fx.micloud.buzz:46001                              CHN      1618    verified          │
│ 1      vless://colonelcy.aurorainiceland.com:443                      USA      2076    verified          │
│ 2      vless://104.17.64.0:2096                                       USA      1580    verified          │
│ 3      vless://104.18.32.0:2096                                       USA      1685    verified          │
│ 4      vless://6234.outline-vpn.cloud:2096                            USA      1342    verified          │
│ 5      vless://104.18.36.81:8080                                      USA      1568    verified          │
│ 6      vless://104.17.96.0:2096                                       USA      1396    verified          │
│ 7      vless://104.19.164.12:8080                                     USA      1162    verified          │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────┘
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
[获取最新的Neobox Key](https://github.com/moqsien/neobox/raw/main/docs/gvc_qq_group.jpg)

**默认本地Inbound**:
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
- [go-qrcode](https://github.com/skip2/go-qrcode)
