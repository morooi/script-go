package script_go

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

func main() {
	secretId := os.Getenv("TENCENTCLOUD_SECRET_ID")
	secretKey := os.Getenv("TENCENTCLOUD_SECRET_KEY")
	certPath := os.Getenv("CERT_PATH")
	keyPath := os.Getenv("KEY_PATH")
	domains := os.Getenv("DOMAINS")

	// Read the certificate and key files
	certContent, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("Failed to read certificate file: %v", err)
	}

	keyContent, err := os.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Failed to read key file: %v", err)
	}

	// Initialize Tencent Cloud client for SSL
	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	sslClient, _ := ssl.NewClient(credential, "", cpf)

	// Get current time for alias
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	alias := fmt.Sprintf("Uploaded at %s", currentTime)

	// Upload the SSL certificate
	uploadRequest := ssl.NewUploadCertificateRequest()
	uploadRequest.CertificatePublicKey = common.StringPtr(string(certContent))
	uploadRequest.CertificatePrivateKey = common.StringPtr(string(keyContent))
	uploadRequest.Alias = common.StringPtr(alias)
	uploadResponse, err := sslClient.UploadCertificate(uploadRequest)
	if err != nil {
		log.Fatalf("Failed to upload certificate: %v", err)
	}

	certId := *uploadResponse.Response.CertificateId
	fmt.Printf("Uploaded certificate ID: %s\n", certId)

	// Initialize Tencent Cloud client for CDN
	cpf.HttpProfile.Endpoint = "cdn.tencentcloudapi.com"
	cdnClient, _ := cdn.NewClient(credential, "", cpf)

	// Split the domains by comma and update each domain
	domainList := strings.Split(domains, ",")
	for _, domain := range domainList {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}

		// Query the current CDN domain configuration
		describeRequest := cdn.NewDescribeDomainsConfigRequest()
		describeRequest.Filters = []*cdn.DomainFilter{
			{
				Name:  common.StringPtr("domain"),
				Value: common.StringPtrs([]string{domain}),
			},
		}

		describeResponse, err := cdnClient.DescribeDomainsConfig(describeRequest)
		if err != nil {
			log.Fatalf("Failed to describe CDN domain %s: %v", domain, err)
		}

		if len(describeResponse.Response.Domains) == 0 {
			log.Fatalf("No CDN configuration found for domain %s", domain)
		}

		currentConfig := describeResponse.Response.Domains[0]

		// Update the CDN domain with the new certificate
		updateRequest := cdn.NewUpdateDomainConfigRequest()
		updateRequest.Domain = common.StringPtr(domain)
		updateRequest.Https = currentConfig.Https
		if updateRequest.Https == nil {
			updateRequest.Https = &cdn.Https{}
		}
		if updateRequest.Https.CertInfo == nil {
			updateRequest.Https.CertInfo = &cdn.ServerCert{}
		}
		updateRequest.Https.CertInfo.CertId = common.StringPtr(certId)

		_, err = cdnClient.UpdateDomainConfig(updateRequest)
		if err != nil {
			log.Fatalf("Failed to update CDN domain %s: %v", domain, err)
		}

		fmt.Printf("CDN domain %s updated successfully\n", domain)
	}
}
