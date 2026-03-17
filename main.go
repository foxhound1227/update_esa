package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

var (
	regionID        = flag.String("region-id", "", "阿里云 Region ID (如 cn-hangzhou)")
	siteID          = flag.Int64("site-id", 0, "ESA 站点 ID")
	configID        = flag.Int64("config-id", 0, "规则配置 ID (优先于 rule-name)")
	ruleName        = flag.String("rule-name", "", "规则名称 (config-id 优先级更高)")
	originScheme    = flag.String("origin-scheme", "", "回源协议 (http, https, follow)")
	httpPort        = flag.Int("http-port", 0, "回源 HTTP 端口")
	httpsPort       = flag.Int("https-port", 0, "回源 HTTPS 端口")
	redirectPort    = flag.Int("redirect-port", 0, "重定向目标端口")
	accessKeyID     = flag.String("access-key-id", "", "阿里云 AccessKey ID (或设置环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID)")
	accessKeySecret = flag.String("access-key-secret", "", "阿里云 AccessKey Secret (或设置环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET)")
	listRules       = flag.Bool("list", false, "列出站点所有规则")
)

func usage() {
	fmt.Println(`ESA 规则更新工具 - 阿里云 ESA 回源/重定向规则管理

用法:
  update_esa [选项]

必需选项:
  --region-id string    阿里云 Region ID (如 cn-hangzhou, cn-shanghai)
  --site-id int64       ESA 站点 ID

规则匹配 (二选一):
  --config-id int64     规则配置 ID (优先)
  --rule-name string    规则名称

操作选项 (至少选择一项):
  # 回源规则操作
  --origin-scheme string    回源协议: http, https, follow
  --http-port int           回源 HTTP 端口
  --https-port int          回源 HTTPS 端口

  # 重定向规则操作
  --redirect-port int       重定向目标端口

  # 其他
  --list                    列出站点所有规则

认证选项:
  --access-key-id string        AccessKey ID
  --access-key-secret string   AccessKey Secret
  # 也可以通过环境变量设置:
  #   export ALIBABA_CLOUD_ACCESS_KEY_ID=xxx
  #   export ALIBABA_CLOUD_ACCESS_KEY_SECRET=xxx

示例:
  # 列出站点所有规则
  update_esa --region-id cn-hangzhou --site-id 123456789 --list

  # 更新回源规则 (使用 config-id)
  update_esa --region-id cn-hangzhou --site-id 123456789 --config-id 473481930420224 --origin-scheme https --https-port 8443

  # 更新回源规则 (使用 rule-name)
  update_esa --region-id cn-hangzhou --site-id 123456789 --rule-name "default" --origin-scheme https --https-port 8443

  # 更新重定向规则端口
  update_esa --region-id cn-hangzhou --site-id 123456789 --rule-name "fn" --redirect-port 8080
  update_esa --region-id cn-hangzhou --site-id 123456789 --config-id 473481930420224 --redirect-port 8080
`)
	flag.PrintDefaults()
}

func buildClient(region string, ak string, sk string) (*esa.Client, error) {
	if ak == "" {
		ak = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	}
	if sk == "" {
		sk = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(ak),
		AccessKeySecret: tea.String(sk),
		RegionId:        tea.String(region),
		Endpoint:        tea.String(fmt.Sprintf("esa.%s.aliyuncs.com", region)),
	}
	return esa.NewClient(config)
}

func getOriginRuleID(client *esa.Client, siteID int64, name string, confID int64) (int64, error) {
	req := &esa.ListOriginRulesRequest{
		SiteId: tea.Int64(siteID),
	}
	res, err := client.ListOriginRules(req)
	if err != nil {
		return 0, err
	}

	configs := res.Body.Configs

	// 优先使用 configID
	if confID != 0 {
		for _, c := range configs {
			if c.ConfigId != nil && *c.ConfigId == confID {
				return confID, nil
			}
		}
		return 0, fmt.Errorf("未找到指定ID的回源规则: %d", confID)
	}

	// 其次使用 ruleName
	if name != "" {
		target := strings.ToLower(strings.TrimSpace(name))
		var names []string
		for _, c := range configs {
			if c.RuleName != nil {
				names = append(names, *c.RuleName)
				if strings.ToLower(strings.TrimSpace(*c.RuleName)) == target {
					return *c.ConfigId, nil
				}
			}
		}
		return 0, fmt.Errorf("未找到指定回源规则名称: %s，可选: %s", name, strings.Join(names, ", "))
	}

	return 0, fmt.Errorf("请指定 --config-id 或 --rule-name")
}

func getRedirectRule(client *esa.Client, siteID int64, name string, confID int64) (*esa.ListRedirectRulesResponseBodyConfigs, error) {
	req := &esa.ListRedirectRulesRequest{
		SiteId: tea.Int64(siteID),
	}
	res, err := client.ListRedirectRules(req)
	if err != nil {
		return nil, err
	}
	configs := res.Body.Configs

	// 优先使用 configID
	if confID != 0 {
		for _, c := range configs {
			if c.ConfigId != nil && *c.ConfigId == confID {
				return c, nil
			}
		}
		return nil, fmt.Errorf("未找到指定ID的重定向规则: %d", confID)
	}

	// 其次使用 ruleName
	if name != "" {
		target := strings.ToLower(strings.TrimSpace(name))
		var names []string
		for _, c := range configs {
			if c.RuleName != nil {
				names = append(names, *c.RuleName)
				if strings.ToLower(strings.TrimSpace(*c.RuleName)) == target {
					return c, nil
				}
			}
		}
		return nil, fmt.Errorf("未找到指定重定向规则名称: %s，可选: %s", name, strings.Join(names, ", "))
	}

	return nil, fmt.Errorf("请指定 --config-id 或 --rule-name")
}

func updateRedirectPortURL(targetURL string, newPort int) string {
	// Pattern to match https://host:port or http://host:port
	patternWithPort := regexp.MustCompile(`(https?://[^:/"]+):(\d+)`)
	if patternWithPort.MatchString(targetURL) {
		return patternWithPort.ReplaceAllString(targetURL, fmt.Sprintf("${1}:%d", newPort))
	}

	// Pattern to match https://host or http://host (no port)
	// We want to replace only the first occurrence if it's not inside another sensitive string
	// But in Go ReplaceAllString replaces all.
	// The python code used count=1.
	// We can use ReplaceAllStringFunc or just FindStringIndex.
	
	patternNoPort := regexp.MustCompile(`(https?://[^:/"]+)`)
	// We only want to replace the first match to be safe, mimicking Python's count=1 which was applied to pattern_no_port
	// Wait, Python's logic:
	// if re.search(pattern_with_port): replace and return
	// else: replace pattern_no_port with count=1
	
	loc := patternNoPort.FindStringIndex(targetURL)
	if loc != nil {
		// loc[0] is start, loc[1] is end
		match := targetURL[loc[0]:loc[1]]
		// match is like "https://host"
		replacement := fmt.Sprintf("%s:%d", match, newPort)
		return targetURL[:loc[0]] + replacement + targetURL[loc[1]:]
	}
	
	return targetURL
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *regionID == "" || *siteID == 0 {
		// Check if it's help
		flag.Usage()
		os.Exit(1)
	}

	client, err := buildClient(*regionID, *accessKeyID, *accessKeySecret)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	if *listRules {
		fmt.Println("=== 回源规则 (Origin Rules) ===")
		originReq := &esa.ListOriginRulesRequest{SiteId: tea.Int64(*siteID)}
		originRes, err := client.ListOriginRules(originReq)
		if err == nil {
			for _, c := range originRes.Body.Configs {
				id := int64(0)
				if c.ConfigId != nil { id = *c.ConfigId }
				name := ""
				if c.RuleName != nil { name = *c.RuleName }
				scheme := ""
				if c.OriginScheme != nil { scheme = *c.OriginScheme }
				httpP := ""
				if c.OriginHttpPort != nil { httpP = *c.OriginHttpPort }
				httpsP := ""
				if c.OriginHttpsPort != nil { httpsP = *c.OriginHttpsPort }
				fmt.Printf("ID: %d\tName: %s\tScheme: %s\tHTTP: %s\tHTTPS: %s\n", id, name, scheme, httpP, httpsP)
			}
		} else {
			fmt.Printf("获取回源规则失败: %v\n", err)
		}

		fmt.Println("\n=== 重定向规则 (Redirect Rules) ===")
		redirectReq := &esa.ListRedirectRulesRequest{SiteId: tea.Int64(*siteID)}
		redirectRes, err := client.ListRedirectRules(redirectReq)
		if err == nil {
			for _, c := range redirectRes.Body.Configs {
				id := int64(0)
				if c.ConfigId != nil { id = *c.ConfigId }
				name := ""
				if c.RuleName != nil { name = *c.RuleName }
				rtype := ""
				if c.Type != nil { rtype = *c.Type }
				target := ""
				if c.TargetUrl != nil { target = *c.TargetUrl }
				fmt.Printf("ID: %d\tName: %s\tType: %s\tTarget: %s\n", id, name, rtype, target)
			}
		} else {
			fmt.Printf("获取重定向规则失败: %v\n", err)
		}
		return
	}

	if *redirectPort != 0 {
		// Update Redirect Rule
		rule, err := getRedirectRule(client, *siteID, *ruleName, *configID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		oldTarget := ""
		if rule.TargetUrl != nil { oldTarget = *rule.TargetUrl }
		newTarget := updateRedirectPortURL(oldTarget, *redirectPort)

		fmt.Printf("正在更新重定向规则: %s (ID: %d)\n", *rule.RuleName, *rule.ConfigId)
		fmt.Printf("原目标: %s\n", oldTarget)
		fmt.Printf("新目标: %s\n", newTarget)

		updateReq := &esa.UpdateRedirectRuleRequest{
			SiteId:    tea.Int64(*siteID),
			ConfigId:  rule.ConfigId,
			TargetUrl: tea.String(newTarget),
		}
		_, err = client.UpdateRedirectRule(updateReq)
		if err != nil {
			fmt.Printf("更新失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("更新成功")

	} else if *originScheme != "" {
		// Update Origin Rule
		cid, err := getOriginRuleID(client, *siteID, *ruleName, *configID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		updateReq := &esa.UpdateOriginRuleRequest{
			SiteId:       tea.Int64(*siteID),
			ConfigId:     tea.Int64(cid),
			OriginScheme: tea.String(*originScheme),
		}
		if *httpPort != 0 {
			updateReq.OriginHttpPort = tea.String(strconv.Itoa(*httpPort))
		}
		if *httpsPort != 0 {
			updateReq.OriginHttpsPort = tea.String(strconv.Itoa(*httpsPort))
		}

		_, err = client.UpdateOriginRule(updateReq)
		if err != nil {
			fmt.Printf("更新回源规则失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("更新回源规则成功")

	} else {
		fmt.Println("错误: 请指定要更新的操作参数 (例如 --origin-scheme 或 --redirect-port)")
		flag.Usage()
		os.Exit(1)
	}
}
