[中文](https://github.com/moqsien/neobox/blob/main/docs/Readme_CN.md)

---------------------------

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
  dedge          Download rawList for a specified edgeTunnel proxy [dedge proxy_index].
  domain         Download selected domains file for edgeTunnels.
  exit           exit the program
  filter         Start filtering proxies by verifier manually.
  gc             Start GC manually.
  geoinfo        Install/Update geoip&geosite for neobox client.
  graw           Manually dowload rawUri list(conf.txt from gitlab) for neobox client.
  guuid          Generate UUIDs.
  help           display help
  parse          Parse rawUri of a proxy to xray-core/sing-box outbound string [xray-core by default].
  pingd          Ping selected domains for edgeTunnels.
  qcode          Generate QRCode for a chosen proxy. [qcode proxy_index]
  remove         Remove a manually added proxy [manually or edgetunnel].
  restart        Restart the running neobox client with a chosen proxy. [restart proxy_index]
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
 args:
  full raw_uri[vless://xxx@xxx?xxx]


>>> restart help

Restart the running neobox client with a chosen proxy. [restart proxy_index]
 options:
  -showchosen; alias:-{sh}; description: show the chosen proxy or not.
  -showconfig; alias:-{shc}; description: show config in result or not.
  -usedomains; alias:-{dom}; description: use selected domains for edgetunnels.
 args:
  choose a specified proxy by index.
```

**Note:**

- **QRCode**: You can use subcommand "qcode" to generate a QRCode for a certain verifed or manually added proxy.
Then you can scan the generated QRCode with NekoBox on your cell phone.
At now, your cell phone will share the proxies verified by neobox.

- **Wireguard**: You can register a wireguard account using subcommand "wireguard". If you specify a Warp+ License Key using "wireguard xxx", 
then you can upgrade your wireguard account to Warp+. Warp+ License Key can be found at Warp+Bot in Telegram App.  The Warp+ License Key is like "Key: 8YK01D7V-Ij6z50g7-eZ5l2R36 (24598562 GB)".

- **Cloudflare EdgeTunnel**: use [pages/workers](https://dash.cloudflare.com/login) to deploy [EdgeTunnel](https://github.com/3Kmfi6HP/EDtunnel). 

- **Cloudflare IP/Domain**: Selected cloudflare IPs/Domains support Warp+ and EdgeTunnel, which can accelerate your visitings through proxies.
  

## Example

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

## Install

```bash
go install -tags "with_wireguard with_shadowsocksr with_utls with_gvisor with_grpc with_ech with_dhcp" github.com/moqsien/neobox/example/neobox@latest
```

Or you can install [gvc](https://github.com/moqsien/gvc) instead.

## Note
You need to **Set A Key**, or the free-vpn-list cannot be decrypted.
Currently, the Key is **5lR3hcN8Zzpo1nzI**.
```bash
>>>setkey 5lR3hcN8Zzpo1nzI

```
[Get Latest Neobox Key](https://github.com/moqsien/neobox/raw/main/docs/gvc_qq_group.jpg)

**Default Local Inbound**: 
```text
http://localhost:2023
```

## Special
**This project is only for study purposes. And no profit is made, neither misleading behaviour is introduced.**
**Chinese users please consciously abide by the laws in China(Mainland).**
**And do not share any info which would be detrimental to the insterests of Our Nation.**


## Thanks to
- [xray-core](https://github.com/XTLS/Xray-core)
- [sing-box](https://github.com/SagerNet/sing-box)
- [wgcf](https://github.com/ViRb3/wgcf)
- [CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest)
- [EDtunnel](https://github.com/3Kmfi6HP/EDtunnel)
- [vpnparser](https://github.com/moqsien/vpnparser)
- [gscraper](https://github.com/moqsien/gscraper)
- [goktrl](https://github.com/moqsien/goktrl)
- [go-qrcode](https://github.com/skip2/go-qrcode)
