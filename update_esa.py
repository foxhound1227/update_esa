import os
import argparse
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

def find_config_id(client, site_id, rule_name=None, config_id=None):
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
        raise RuntimeError("未找到指定规则名称: " + rule_name + " 可选: " + (", ".join(names) if names else "无"))
    if configs:
        return configs[0].config_id
    raise RuntimeError("站点下无回源规则配置")

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--region-id", required=True)
    parser.add_argument("--site-id", required=True)
    parser.add_argument("--config-id")
    parser.add_argument("--rule-name")
    parser.add_argument("--origin-scheme", choices=["http", "https", "follow"], required=True)
    parser.add_argument("--http-port", type=int)
    parser.add_argument("--https-port", type=int)
    parser.add_argument("--access-key-id")
    parser.add_argument("--access-key-secret")
    parser.add_argument("--list", action="store_true")
    args = parser.parse_args()

    client = build_client(args.region_id, args.access_key_id, args.access_key_secret)
    if args.list:
        res = client.list_origin_rules(esa_models.ListOriginRulesRequest(site_id=int(args.site_id)))
        for c in res.body.configs or []:
            print(str(c.config_id) + "\t" + str(c.rule_name) + "\t" + str(c.origin_scheme) + "\t" + str(c.origin_http_port) + "\t" + str(c.origin_https_port))
        return
    cid = find_config_id(client, args.site_id, args.rule_name, args.config_id)
    req = esa_models.UpdateOriginRuleRequest(
        site_id=int(args.site_id),
        config_id=int(cid),
        origin_scheme=args.origin_scheme,
        origin_http_port=args.http_port if args.http_port else None,
        origin_https_port=args.https_port if args.https_port else None,
    )
    res = client.update_origin_rule(req)
    print(res.body.request_id)

if __name__ == "__main__":
    main()
