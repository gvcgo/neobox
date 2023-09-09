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
  --port=xxx; alias:-{p}; description: port for edge tunnel.
 args:
  full raw_uri[vless://xxx@xxx?xxx]
```

## 在Shell中查看Neobox的运行状态

```bash

>>> show
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
|                                                                                                     |
| RawList[5381@2023-09-09 11:53:13] vmess[2752] vless[317] trojan[602] ss[1638] ssr[72]               |
| Pinged[457@2023-09-09 20:50:44] vmess[291] vless[23] trojan[69] ss[73] ssr[1]                       |
| Pinged[10@2023-09-09 20:52:29] vmess[4] vless[2] trojan[4] ss[0] ssr[0]                             |
| Database: History[6] EdgeTunnel[1] Manually[0]                                                      |
| idx proxy                                                                   location rtt(ms) source |
| 0 vmess://jp6.lianpi.xyz:23234                                             USA  1278 verified       |
| 1 vmess://172.67.83.191:80                                                 USA  2022 verified       |
| 2 vmess://cdn.chigua.tk:80                                                 CHN  2193 verified       |
| 3 vmess://46caadb.f2.gladns.com:3331                                       USA  1519 verified       |
| 4 vless://172.66.47.84:443                                                 USA  1589 verified       |
| 5 vless://realtu.skrspc.com:50339                                          JPN  1374 verified       |
| 6 trojan://18.198.10.5:22222                                               DEU  1875 verified       |
| 7 trojan://13.50.193.134:22222                                             USA  1403 verified       |
| 8 trojan://20.212.51.48:443                                                USA  2043 verified       |
| 9 trojan://1gzdx2.156786.xyz:37905                                         JPN  2456 verified       |
| w0 wireguard://172.64.152.234:2053                                         USA  175  wireguard      |
| e0 vless://xx-edtunnel-xxx.xxxxx.dev:443                                   USA  250  edtunnel       |
| NeoBox[running @ (Mem: 4220MiB)] Verifier[completed] Keeper[running]                                |
| LogFileDir: C:\home\.neobox\log_files                                                               |
|                                                                                                     |
|                                                                                                     |
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## 如何安装？

```bash
go install github.com/moqsien/neobox/example/neobox@latest
```

## 注意事项
想要使用免费代理，则必须获取解密Key. 当前用于解密的Key是 **5lR3hcN8Zzpo1nzI**.

设置解密Key：
```bash
>>>setkey 5lR3hcN8Zzpo1nzI

```

## 特别声明
本项目仅作学习研究使用。无意违反任何国家法律法规。请使用者谨慎使用。

## 感谢
- [xray-core](https://github.com/XTLS/Xray-core)
- [sing-box](https://github.com/SagerNet/sing-box)
- [wgcf](https://github.com/ViRb3/wgcf)
- [CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest)
- [EDtunnel](https://github.com/3Kmfi6HP/EDtunnel)
- [vpnparser](https://github.com/moqsien/vpnparser)
- [gscraper](https://github.com/moqsien/gscraper)
