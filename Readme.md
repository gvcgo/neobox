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
  exit           exit the program
  filter         Start filtering proxies by verifier manually.
  gc             Start GC manually.
  geoinfo        Install/Update geoip&geosite for neobox client.
  graw           Manually dowload rawUri list(conf.txt from gitlab) for neobox client.
  guuid          Generate UUIDs.
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
 args:
  full raw_uri[vless://xxx@xxx?xxx]
```

## Example

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

Defaul Local Inbound: 
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
