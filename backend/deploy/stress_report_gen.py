#!/usr/bin/env python3
"""Generate Markdown stress test report. Args: see main()."""
import sys
from collections import defaultdict

# TSV stage -> HTTP API（用于 QPS；含成功与失败请求）
API_BY_STAGE = {
    ("PAYIN", "create"): "POST /v1/payin/order",
    ("PAYIN", "notify"): "POST /v1/callback/notify",
    ("PAYIN", "query"): "GET /v1/payin/query",
    ("PAYOUT", "create"): "POST /v1/payout/order",
    ("PAYOUT", "query"): "GET /v1/payout/query",
}


def main() -> None:
    if len(sys.argv) < 20:
        print(
            "usage: stress_report_gen.py RESULTS started ended elapsed inv_rc gw merchant "
            "iters_config duration_sec actual_rounds conc payin_w pam pom payin_type "
            "skip_notify skip_inv inv_out_path report_path",
            file=sys.stderr,
        )
        sys.exit(2)
    (
        results_path,
        started_at,
        ended_at,
        elapsed_s,
        inv_rc_s,
        gw,
        merchant,
        iters_config,
        duration_sec,
        actual_rounds,
        conc,
        payin_w,
        pam,
        pom,
        payin_type,
        skip_notify,
        skip_inv,
        inv_out_path,
        report_path,
    ) = sys.argv[1:20]
    elapsed = max(int(elapsed_s), 0)
    inv_rc = int(inv_rc_s)

    counts = defaultdict(int)
    lat_sum = defaultdict(float)
    lat_n = defaultdict(int)
    http_bucket = defaultdict(int)
    endpoint_hits = defaultdict(int)

    try:
        with open(results_path, encoding="utf-8") as f:
            lines = [ln.strip() for ln in f if ln.strip()]
    except FileNotFoundError:
        lines = []

    for ln in lines:
        parts = ln.split("\t")
        if len(parts) < 6:
            continue
        op, stage, ok, http, t, _extra = parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]
        counts[f"{op}:{stage}:{ok}"] += 1
        label = API_BY_STAGE.get((op, stage))
        if label:
            endpoint_hits[label] += 1
        try:
            tf = float(t)
            lat_sum[op + ":" + stage] += tf
            lat_n[op + ":" + stage] += 1
        except ValueError:
            pass
        http_bucket[f"{op}:{http}"] += 1

    def avg(op_stage: str) -> float:
        n = lat_n.get(op_stage, 0)
        return 0.0 if n == 0 else lat_sum[op_stage] / n

    inv_body = open(inv_out_path, encoding="utf-8", errors="replace").read()

    denom = float(elapsed) if elapsed > 0 else 0.0
    total_http = sum(endpoint_hits.values())
    overall_qps = (total_http / denom) if denom > 0 else 0.0

    out: list[str] = []
    out.append("# 代收 / 代付 OpenAPI 压力测试报告\n")
    out.append(f"- 开始: `{started_at}`\n")
    out.append(f"- 结束: `{ended_at}`\n")
    out.append(f"- 压力阶段墙钟: **{elapsed}s**\n")
    out.append(f"- 网关: `{gw}`\n")
    out.append(f"- 商户: `{merchant}`\n\n")

    out.append("## 配置\n\n")
    out.append("| 项 | 值 |\n|---|---|\n")
    out.append(f"| STRESS_DURATION_SEC | {duration_sec}（0 表示按轮次） |\n")
    out.append(f"| STRESS_ITERATIONS（配置） | {iters_config} |\n")
    out.append(f"| 实际完成轮次 | {actual_rounds} |\n")
    out.append(f"| STRESS_CONCURRENCY | {conc} |\n")
    out.append(f"| PAYIN_WEIGHT_PERCENT | {payin_w} |\n")
    out.append(f"| PAYIN_TYPE / PAYIN_AMOUNT | {payin_type} / {pam} 分 |\n")
    out.append(f"| PAYOUT_AMOUNT | {pom} 分 |\n")
    out.append(f"| STRESS_SKIP_PAYIN_NOTIFY | {skip_notify} |\n")
    out.append(f"| STRESS_SKIP_INVARIANTS | {skip_inv} |\n\n")

    out.append("## 接口 QPS（压力阶段内，按请求次数 / 墙钟秒；含失败请求）\n\n")
    out.append(
        f"- **压力阶段 HTTP 总请求数**: {total_http}（create/notify/query 合计）\n"
    )
    if denom > 0:
        out.append(f"- **总 QPS（上述合计）**: **{overall_qps:.2f}**\n\n")
    else:
        out.append("- **总 QPS**: N/A（elapsed=0）\n\n")

    out.append("| 接口 | 请求数 | QPS |\n")
    out.append("|---|---:|---:|\n")
    order = [
        "POST /v1/payin/order",
        "POST /v1/callback/notify",
        "GET /v1/payin/query",
        "POST /v1/payout/order",
        "GET /v1/payout/query",
    ]
    for api in order:
        n = endpoint_hits.get(api, 0)
        if n == 0:
            continue
        q = (n / denom) if denom > 0 else 0.0
        out.append(f"| `{api}` | {n} | {q:.2f} |\n")
    for api, n in sorted(endpoint_hits.items()):
        if api in order:
            continue
        q = (n / denom) if denom > 0 else 0.0
        out.append(f"| `{api}` | {n} | {q:.2f} |\n")
    out.append("\n")

    out.append("## 结果汇总（按阶段）\n\n")
    out.append("| 业务 | 阶段 | 成功 | 失败 |\n")
    out.append("|---|---|---:|---:|\n")
    for op in ("PAYIN", "PAYOUT"):
        for stage in ("create", "notify", "query"):
            ok = counts.get(f"{op}:{stage}:OK", 0)
            fail = counts.get(f"{op}:{stage}:FAIL", 0)
            if ok == 0 and fail == 0:
                continue
            out.append(f"| {op} | {stage} | {ok} | {fail} |\n")

    out.append("\n## HTTP 状态分布\n\n")
    out.append("| 组合 | 次数 |\n")
    out.append("|---|---:|\n")
    for k in sorted(http_bucket.keys()):
        out.append(f"| `{k}` | {http_bucket[k]} |\n")

    out.append("\n## 平均耗时（秒, curl time_total）\n\n")
    out.append("| 阶段 | avg(s) | 样本 |\n")
    out.append("|---|---:|---:|\n")
    for op in ("PAYIN", "PAYOUT"):
        for stage in ("create", "query"):
            k = f"{op}:{stage}"
            n = lat_n.get(k, 0)
            if n == 0:
                continue
            out.append(f"| {k} | {avg(k):.4f} | {n} |\n")

    out.append("\n## DB 不变量检查\n\n")
    out.append(f"- **退出码**: `{inv_rc}`（0 表示通过）\n\n")
    out.append("```text\n")
    out.append(inv_body)
    out.append("\n```\n\n")

    out.append("## 结论\n\n")
    fail_total = sum(
        counts.get(f"{op}:{s}:FAIL", 0)
        for op in ("PAYIN", "PAYOUT")
        for s in ("create", "notify", "query")
    )
    if skip_inv == "1":
        if fail_total == 0:
            out.append("- 本次样本：**接口阶段无失败**；**未运行** DB 不变量（`STRESS_SKIP_INVARIANTS=1`）。\n")
        else:
            out.append("- 接口阶段存在失败；未运行 DB 不变量。请结合上表排查。\n")
    elif inv_rc == 0 and fail_total == 0:
        out.append("- 本次样本：**接口阶段无失败**且 **DB 不变量通过**。\n")
    else:
        out.append("- 存在失败计数或不变量未通过，请结合上表与 `reason_code` / `code`、日志排查。\n")
    out.append(f"- 原始 TSV 行数：**{len(lines)}**。\n")

    open(report_path, "w", encoding="utf-8").write("".join(out))


if __name__ == "__main__":
    main()
