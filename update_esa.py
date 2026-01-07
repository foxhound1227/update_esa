import os
import argparse
import re
from alibabacloud_tea_openapi import models as open_api_models
from alibabacloud_esa20240910.client import Client as EsaClient
from alibabacloud_esa20240910 import models as esa_models

def build_client(region_id, access_key_id=None, access_key_secret=None):
    config = open_api_models.Config(
        access_key_id=access_key_id or os.getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
        access_key_secret=access_key_secret or os.getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
        region_id=region_id,
        endpoint=f"esa.{region_id}.aliyuncs.com"
    )
    return EsaClient(config)

def get_origin_rule_id(client, site_id, rule_name=None, config_id=None):
    if config_id:
        return int(config_id)
    req = esa_models.ListOriginRulesRequest(site_id=int(site_id))
    res = client.list_origin_rules(req)
    configs = (res.body.configs or [])
    if rule_name:
        target = (rule_name or "").strip().lower()
        for c in configs:
            if ((c.rule_name or "").strip().lower()) == target:
                return c.config_id
        names = [c.rule_name for c in configs if c.rule_name]
        raise RuntimeError("未找到指定回源规则名称: " + rule_name + " 可选: " + (", ".join(names) if names else "无"))
    if configs:
        return configs[0].config_id
    raise RuntimeError("站点下无回源规则配置")

def get_redirect_rule(client, site_id, rule_name=None, config_id=None):
    req = esa_models.ListRedirectRulesRequest(site_id=int(site_id))
    res = client.list_redirect_rules(req)
    configs = (res.body.configs or [])
    
    if config_id:
        cid = int(config_id)
        for c in configs:
            if c.config_id == cid:
                return c
        raise RuntimeError(f"未找到指定ID的重定向规则: {config_id}")
        
    if rule_name:
        target = (rule_name or "").strip().lower()
        for c in configs:
            if ((c.rule_name or "").strip().lower()) == target:
                return c
        names = [c.rule_name for c in configs if c.rule_name]
        raise RuntimeError("未找到指定重定向规则名称: " + rule_name + " 可选: " + (", ".join(names) if names else "无"))
        
    if configs:
        return configs[0]
    raise RuntimeError("站点下无重定向规则配置")

def update_redirect_port(target_url, new_port):
    # Pattern to match https://host:port or https://host
    # It handles simple strings or inside concat("...", ...)
    
    # Check if port exists
    pattern_with_port = r'(https?://[^:/"]+):(\d+)'
    
    if re.search(pattern_with_port, target_url):
        return re.sub(pattern_with_port, f'\\1:{new_port}', target_url)
    
    # If no port, append it to the host
    pattern_no_port = r'(https?://[^:/"]+)'
    # Be careful not to match if it's already handled, but above check covers it.
    # We need to make sure we don't match inside other structures inappropriately.
    # The target_url in ESA dynamic redirect usually is:
    # concat("https://host:port", ...)
    
    return re.sub(pattern_no_port, f'\\1:{new_port}', target_url, count=1)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--region-id", required=True)
    parser.add_argument("--site-id", required=True)
    parser.add_argument("--config-id")
    parser.add_argument("--rule-name")
    
    # Origin Rule params
    parser.add_argument("--origin-scheme", choices=["http", "https", "follow"])
    parser.add_argument("--http-port", type=int)
    parser.add_argument("--https-port", type=int)
    
    # Redirect Rule params
    parser.add_argument("--redirect-port", type=int, help="修改重定向规则的目标端口")
    
    parser.add_argument("--access-key-id")
    parser.add_argument("--access-key-secret")
    parser.add_argument("--list", action="store_true")
    args = parser.parse_args()

    client = build_client(args.region_id, args.access_key_id, args.access_key_secret)
    
    if args.list:
        print("=== 回源规则 (Origin Rules) ===")
        try:
            res = client.list_origin_rules(esa_models.ListOriginRulesRequest(site_id=int(args.site_id)))
            for c in res.body.configs or []:
                print(f"ID: {c.config_id}\tName: {c.rule_name}\tScheme: {c.origin_scheme}\tHTTP: {c.origin_http_port}\tHTTPS: {c.origin_https_port}")
        except Exception as e:
            print(f"获取回源规则失败: {e}")
            
        print("\n=== 重定向规则 (Redirect Rules) ===")
        try:
            res = client.list_redirect_rules(esa_models.ListRedirectRulesRequest(site_id=int(args.site_id)))
            for c in res.body.configs or []:
                print(f"ID: {c.config_id}\tName: {c.rule_name}\tType: {c.type}\tTarget: {c.target_url}")
        except Exception as e:
            print(f"获取重定向规则失败: {e}")
        return

    # Dispatch based on arguments
    if args.redirect_port:
        # Update Redirect Rule
        rule = get_redirect_rule(client, args.site_id, args.rule_name, args.config_id)
        new_target_url = update_redirect_port(rule.target_url, args.redirect_port)
        
        print(f"正在更新重定向规则: {rule.rule_name} (ID: {rule.config_id})")
        print(f"原目标: {rule.target_url}")
        print(f"新目标: {new_target_url}")
        
        req = esa_models.UpdateRedirectRuleRequest(
            site_id=int(args.site_id),
            config_id=rule.config_id,
            target_url=new_target_url
        )
        res = client.update_redirect_rule(req)
        print(f"更新成功, RequestId: {res.body.request_id}")
        
    elif args.origin_scheme:
        # Update Origin Rule
        cid = get_origin_rule_id(client, args.site_id, args.rule_name, args.config_id)
        req = esa_models.UpdateOriginRuleRequest(
            site_id=int(args.site_id),
            config_id=int(cid),
            origin_scheme=args.origin_scheme,
            origin_http_port=args.http_port if args.http_port else None,
            origin_https_port=args.https_port if args.https_port else None,
        )
        res = client.update_origin_rule(req)
        print(f"更新回源规则成功, RequestId: {res.body.request_id}")
    else:
        print("错误: 请指定要更新的操作参数 (例如 --origin-scheme 或 --redirect-port)")

if __name__ == "__main__":
    main()
