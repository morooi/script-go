#!/bin/sh
go build -o update_cert update_tencent_cloud_cdn_cert.go
go build -o refresh_cdn update_tencent_cloud_cdn_cert.go
go build -o audit_time_analyze audit_time_analyze.go
