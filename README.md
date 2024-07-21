## GO 脚本

### 上传 SSL 证书到腾讯云并更新腾讯云 CDN

脚本： [update_tencent_cloud_cdn_cert.go](update_tencent_cloud_cdn_cert.go)

编译命令：

```shell
GOOS=linux GOARCH=amd64 go build -o update_cert update_tencent_cloud_cdn_cert.go
```

acme.sh 命令：

```shell
acme.sh --install-cert -d your.domain.com \
   --key-file /path/to/key.key \
   --fullchain-file /path/to/fullchain.cer \
   --reloadcmd "TENCENTCLOUD_SECRET_ID=YOUR_SECRET_ID TENCENTCLOUD_SECRET_KEY=YOUR_SECRET_KEY CERT_PATH=/path/to/fullchain.cer KEY_PATH=/path/to/key.key DOMAINS='domain1.com,domain2.com,domain3.com' /path/to/update_cert"
```




