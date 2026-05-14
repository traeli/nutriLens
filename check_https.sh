#!/bin/bash

DOMAIN="nutrilens.cloud"
API_PATH="/api/v1/auth/wx-login"

echo "=============================="
echo "1. 检查证书文件"
echo "=============================="
CERT="/etc/openresty/certificate/nutrilens.cloud_bundle.pem"
KEY="/etc/openresty/certificate/nutrilens.cloud.key"

if [[ -f "$CERT" ]]; then
    echo "证书存在: $CERT"
else
    echo "证书不存在！"
fi

if [[ -f "$KEY" ]]; then
    echo "私钥存在: $KEY"
else
    echo "私钥不存在！"
fi

echo ""
echo "=============================="
echo "2. 检查 TLS 握手 (openssl s_client)"
echo "=============================="
for TLS_VER in tls1_2 tls1_3; do
    echo "尝试 TLS version: $TLS_VER"
    openssl s_client -connect $DOMAIN:443 -$TLS_VER -servername $DOMAIN < /dev/null \
        2>/dev/null | grep -E 'Protocol|Cipher'
done

echo ""
echo "=============================="
echo "3. 检查 HTTP/2 支持 (curl)"
echo "=============================="
curl -vk --http2 https://$DOMAIN/

echo ""
echo "=============================="
echo "4. 检查 API 接口"
echo "=============================="
curl -vk --http2 https://$DOMAIN$API_PATH