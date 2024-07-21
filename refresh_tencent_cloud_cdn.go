package main

import (
	"fmt"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func main() {
	secretId := os.Getenv("TENCENTCLOUD_SECRET_ID")
	secretKey := os.Getenv("TENCENTCLOUD_SECRET_KEY")
	refreshPath := os.Getenv("CDN_REFRESH_PATH")
	if secretId == "" || secretKey == "" || refreshPath == "" {
		fmt.Println("请设置环境变量 TENCENTCLOUD_SECRET_ID, TENCENTCLOUD_SECRET_KEY 和 CDN_REFRESH_PATH")
		os.Exit(1)
	}

	// 配置API客户端
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cdn.tencentcloudapi.com"

	// 创建CDN客户端
	client, err := cdn.NewClient(credential, "", cpf)
	if err != nil {
		fmt.Printf("failed to create cdn client: %v\n", err)
		os.Exit(1)
	}

	// 创建刷新请求
	request := cdn.NewPurgePathCacheRequest()
	request.Paths = common.StringPtrs([]string{
		refreshPath,
	})
	request.FlushType = common.StringPtr("delete")

	// 发送请求
	response, err := client.PurgePathCache(request)
	if err != nil {
		fmt.Printf("failed to refresh cdn cache: %v\n", err)
		os.Exit(1)
	}

	// 打印请求ID
	fmt.Printf("Request ID: %s\n", *response.Response.RequestId)
}
