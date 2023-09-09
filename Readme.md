## What's neobox?

Neobox is a powerful free vpn manager.
It supports protocols like vmess/vless/trojan/shadowsocks/shadowsocksr/wireguard.

## What can neobox do?

- Test free vpns.
- Test cloudflare CDN IPs.
- Cloudflare EdgeTunnel.
- Cloudflare wireguard/warp-plus.
- An interactive shell providing all the features above.

## Shell command

```bash
$neobox shell
>>> help

Commands:
  add            Add proxies to neobox mannually.
  added          Add edgetunnel proxies to neobox.
  cfip           Test speed for cloudflare IPv4s.
  clear          clear the screen
  exit           exit the program
  filter         Start filtering proxies by verifier manually.
  gc             Start GC manually.
  geoinfo        Install/Update geoip&geosite for neobox client.
  graw           Manually dowload rawUri list(conf.txt from gitlab) for neobox client.
  help           display help
  restart        Restart the running neobox client with a chosen proxy. [restart proxy_index]
  rmproxy        Remove a manually added proxy [manually or edgetunnel].
  setkey         Setup rawlist encrytion key for neobox. [With no args will set key to default value]
  setping        Setup ping without root for Linux.
  show           Show neobox info.
  start          Start a neobox client/keeper.
  stop           Stop neobox client.
  wireguard      Register wireguard account and update licenseKey to Warp+ [if a licenseKey is specified].


>>> added help

Add edgetunnel proxies to neobox.
 options:
  --uuid=xxx; alias:-{u}; description: uuid for edge tunnel vless.
  --address=xxx; alias:-{a}; description: domain/ip for edge tunnel.
  --port=xxx; alias:-{p}; description: port for edge tunnel.
 args:
  full raw_uri[vless://xxx@xxx?xxx]
```

## Example

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

## Install

```bash
go install github.com/moqsien/neobox/example/neobox@latest
```

## Note
You need to **Set A Key**, or the free-vpn-list cannot be decrypted.
Currently, the Key is **5lR3hcN8Zzpo1nzI**.
```bash
>>>setkey 5lR3hcN8Zzpo1nzI

```

## Special
This project is only for study purposes.

## Thanks to
- [xray-core](https://github.com/XTLS/Xray-core)
- [sing-box](https://github.com/SagerNet/sing-box)
- [wgcf](https://github.com/ViRb3/wgcf)
- [CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest)
- [EDtunnel](https://github.com/3Kmfi6HP/EDtunnel)
- [vpnparser](https://github.com/moqsien/vpnparser)
- [gscraper](https://github.com/moqsien/gscraper)
