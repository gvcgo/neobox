#!/bin/bash
go build -tags "with_wireguard with_shadowsocksr with_utls with_gvisor with_grpc with_ech with_dhcp" -o nbox ./neobox
