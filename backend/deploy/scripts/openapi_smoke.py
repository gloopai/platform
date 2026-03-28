#!/usr/bin/env python3
"""
OpenAPI 联调脚本：代收/代付下单与查单、余额（gateway OpenAPIServer，默认 :8090）。

- 签名与 gateway internal/middleware/sign_md5.go 的 Md5Sign 一致。
- 业务成功以统一 JSON 信封 code=2000 为准（与 apiresp.CodeSuccess 一致）。

主入口（一条命令跑通全流程，订单号脚本内随机生成）：

  python3 openapi_smoke.py
  python3 openapi_smoke.py run

单步子命令仍可用：payin-create / payout-create / payin-query / payout-query / balance / check
"""

from __future__ import annotations

import argparse
import hashlib
import http.client
import json
import secrets
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any

API_CODE_SUCCESS = 2000


def md5_sign(params: dict[str, str], secret: str) -> str:
    keys = sorted(k.lower() for k in params if k.lower() != "sign")
    parts: list[str] = []
    for k in keys:
        v = (params.get(k) or "").strip()
        if not v:
            continue
        parts.append(f"{k}={v}")
    base = "&".join(parts)
    if base:
        base += "&"
    base += f"key={secret}"
    return hashlib.md5(base.encode("utf-8")).hexdigest()


def random_merchant_order_no(prefix: str) -> str:
    """商户侧唯一订单号（时间戳 + 随机 hex，避免联调重复）。"""
    return f"{prefix}-{int(time.time() * 1000)}-{secrets.token_hex(4)}"


def _with_sign(
    base: dict[str, str],
    *,
    app_id: str,
    api_secret: str,
) -> dict[str, str]:
    p = {k.lower(): str(v).strip() for k, v in base.items() if v is not None}
    p["app_id"] = app_id
    p["timestamp"] = str(int(time.time()))
    p["nonce"] = secrets.token_hex(8)
    p["sign"] = md5_sign(p, api_secret)
    return p


def _conn_error_hint(url: str, err: BaseException) -> str:
    health = urllib.parse.urljoin(url, "/health")
    return (
        f"HTTP 连接失败: {err!r}\n"
        f"  目标: {url}\n"
        "  常见原因: 本机未启动 gateway，或 OpenAPI 不是 8090；先执行:\n"
        f"    curl -sS '{health}'\n"
        "  应能返回 JSON；若失败请在 gateway 目录用 -f etc/gateway-api.yaml 启动（四端口含 OpenAPIServer）。\n"
    )


def _post_json(url: str, body: dict[str, Any], timeout: float = 30.0) -> tuple[int, str]:
    data = json.dumps(body, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        headers={
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Connection": "close",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.getcode(), resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8")
    except (urllib.error.URLError, http.client.RemoteDisconnected, ConnectionResetError, BrokenPipeError, OSError) as e:
        sys.stderr.write(_conn_error_hint(url, e))
        raise SystemExit(1) from None


def _get(url: str, timeout: float = 30.0) -> tuple[int, str]:
    req = urllib.request.Request(url, headers={"Accept": "application/json", "Connection": "close"}, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.getcode(), resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8")
    except (urllib.error.URLError, http.client.RemoteDisconnected, ConnectionResetError, BrokenPipeError, OSError) as e:
        sys.stderr.write(_conn_error_hint(url, e))
        raise SystemExit(1) from None


def parse_envelope(text: str) -> tuple[bool, int | None, str, dict[str, Any]]:
    """返回 (业务是否成功, code, message, data)。"""
    try:
        o = json.loads(text)
    except json.JSONDecodeError:
        return False, None, "响应不是合法 JSON", {}
    if not isinstance(o, dict):
        return False, None, "响应不是 JSON 对象", {}
    code = o.get("code")
    if not isinstance(code, int):
        return False, None, "缺少或非法 code 字段", {}
    msg = str(o.get("message") or "")
    raw = o.get("data")
    data: dict[str, Any] = raw if isinstance(raw, dict) else {}
    return code == API_CODE_SUCCESS, code, msg, data


def is_api_success(http_code: int, text: str) -> tuple[bool, int | None, str, dict[str, Any]]:
    if http_code != 200:
        return False, None, f"HTTP {http_code}", {}
    ok, code, msg, data = parse_envelope(text)
    if not ok:
        return False, code, msg, data
    return True, code, msg, data


def _signed_flat_to_payin_json(flat: dict[str, str]) -> dict[str, Any]:
    out: dict[str, Any] = {
        "app_id": flat["app_id"],
        "merchant_order_no": flat["merchant_order_no"],
        "amount": int(flat["amount"]),
        "currency": flat["currency"],
        "timestamp": int(flat["timestamp"]),
        "nonce": flat["nonce"],
        "sign": flat["sign"],
    }
    if flat.get("payin_type"):
        out["payin_type"] = flat["payin_type"]
    if flat.get("notify_url"):
        out["notify_url"] = flat["notify_url"]
    if flat.get("return_url"):
        out["return_url"] = flat["return_url"]
    if flat.get("subject"):
        out["subject"] = flat["subject"]
    return out


def _signed_flat_to_payout_json(flat: dict[str, str]) -> dict[str, Any]:
    out: dict[str, Any] = {
        "app_id": flat["app_id"],
        "merchant_order_no": flat["merchant_order_no"],
        "amount": int(flat["amount"]),
        "currency": flat["currency"],
        "timestamp": int(flat["timestamp"]),
        "nonce": flat["nonce"],
        "sign": flat["sign"],
        "payout_product_code": flat["payout_product_code"],
    }
    if flat.get("notify_url"):
        out["notify_url"] = flat["notify_url"]
    return out


def _query_qs(params: dict[str, str], app_id: str, secret: str) -> str:
    p = _with_sign(params, app_id=app_id, api_secret=secret)
    return urllib.parse.urlencode(p)


def do_payin_create(
    base: str,
    app_id: str,
    secret: str,
    *,
    merchant_order_no: str,
    amount: int,
    currency: str,
    payin_type: str,
    notify_url: str = "",
    return_url: str = "",
    subject: str = "",
) -> tuple[int, str]:
    b = {
        "merchant_order_no": merchant_order_no,
        "amount": str(amount),
        "currency": currency,
        "payin_type": payin_type,
        "notify_url": notify_url,
        "return_url": return_url,
        "subject": subject,
    }
    flat = {k: v for k, v in _with_sign(b, app_id=app_id, api_secret=secret).items() if v != ""}
    body = _signed_flat_to_payin_json(flat)
    url = base.rstrip("/") + "/v1/payin/order"
    return _post_json(url, body)


def do_payout_create(
    base: str,
    app_id: str,
    secret: str,
    *,
    merchant_order_no: str,
    amount: int,
    currency: str,
    payout_product_code: str,
    notify_url: str = "",
) -> tuple[int, str]:
    b = {
        "merchant_order_no": merchant_order_no,
        "amount": str(amount),
        "currency": currency,
        "notify_url": notify_url,
        "payout_product_code": payout_product_code,
    }
    flat = {k: v for k, v in _with_sign(b, app_id=app_id, api_secret=secret).items() if v != ""}
    body = _signed_flat_to_payout_json(flat)
    url = base.rstrip("/") + "/v1/payout/order"
    return _post_json(url, body)


def do_payin_query(base: str, app_id: str, secret: str, *, order_no: str = "", merchant_order_no: str = "") -> tuple[int, str]:
    q: dict[str, str] = {}
    if order_no:
        q["order_no"] = order_no
    if merchant_order_no:
        q["merchant_order_no"] = merchant_order_no
    qs = _query_qs(q, app_id, secret)
    url = base.rstrip("/") + "/v1/payin/query?" + qs
    return _get(url)


def do_payout_query(base: str, app_id: str, secret: str, *, order_no: str = "", merchant_order_no: str = "") -> tuple[int, str]:
    q: dict[str, str] = {}
    if order_no:
        q["order_no"] = order_no
    if merchant_order_no:
        q["merchant_order_no"] = merchant_order_no
    qs = _query_qs(q, app_id, secret)
    url = base.rstrip("/") + "/v1/payout/query?" + qs
    return _get(url)


def do_balance(base: str, app_id: str, secret: str) -> tuple[int, str]:
    qs = _query_qs({}, app_id, secret)
    url = base.rstrip("/") + "/v1/merchant/balance/query?" + qs
    return _get(url)


def cmd_run_all(args: argparse.Namespace) -> int:
    base = args.base.rstrip("/")
    app_id = args.app_id
    secret = args.secret
    lines: list[str] = []
    failed = False

    def step(name: str, ok: bool, detail: str = "") -> None:
        nonlocal failed
        mark = "PASS" if ok else "FAIL"
        if not ok:
            failed = True
        line = f"  [{mark}] {name}"
        if detail:
            line += f" — {detail}"
        lines.append(line)

    # 1) health
    hc, ht = _get(base + "/health")
    ok, code, msg, _ = is_api_success(hc, ht)
    step("GET /health", ok, f"code={code} {msg}" if code is not None else ht[:200])

    # 2) balance (before)
    bc, bt = do_balance(base, app_id, secret)
    ok, code, msg, bdata = is_api_success(bc, bt)
    bal_detail = ""
    if ok and bdata:
        bal_detail = f"available={bdata.get('available_balance')} payin={bdata.get('payin_balance')}"
    step("GET /v1/merchant/balance/query (before)", ok, f"{msg} {bal_detail}".strip())

    # 3) payin create
    mo = random_merchant_order_no("MO")
    pc, pt = do_payin_create(
        base,
        app_id,
        secret,
        merchant_order_no=mo,
        amount=args.payin_amount,
        currency=args.currency,
        payin_type=args.payin_type,
    )
    ok, code, msg, pdata = is_api_success(pc, pt)
    order_no_in = str(pdata.get("order_no") or "") if pdata else ""
    step(
        "POST /v1/payin/order",
        ok,
        f"merchant_order_no={mo} order_no={order_no_in} {msg}".strip(),
    )
    if not ok or not order_no_in:
        lines.append("  (后续代收查单跳过：下单未成功或未返回 order_no)")
        payin_ok = False
    else:
        payin_ok = True

    # 4) payin query
    if payin_ok:
        qc, qt = do_payin_query(base, app_id, secret, order_no=order_no_in)
        ok, code, msg, qdata = is_api_success(qc, qt)
        on2 = ""
        if ok and qdata.get("order"):
            on2 = str((qdata["order"] or {}).get("order_no") or "")
        step("GET /v1/payin/query", ok, f"order_no={on2 or order_no_in} {msg}".strip())

    # 5) payout create
    mp = random_merchant_order_no("MP")
    poc, pot = do_payout_create(
        base,
        app_id,
        secret,
        merchant_order_no=mp,
        amount=args.payout_amount,
        currency=args.currency,
        payout_product_code=args.payout_product_code,
    )
    ok, code, msg, pod = is_api_success(poc, pot)
    po_no = str(pod.get("order_no") or "") if pod else ""
    step(
        "POST /v1/payout/order",
        ok,
        f"merchant_order_no={mp} order_no={po_no} {msg}".strip(),
    )
    payout_ok = ok and bool(po_no)

    # 6) payout query
    if payout_ok:
        pqc, pqt = do_payout_query(base, app_id, secret, order_no=po_no)
        ok, code, msg, pqdata = is_api_success(pqc, pqt)
        step("GET /v1/payout/query", ok, msg)

    # 7) balance (after)
    ac, at = do_balance(base, app_id, secret)
    ok, code, msg, adata = is_api_success(ac, at)
    bal2 = ""
    if ok and adata:
        bal2 = f"available={adata.get('available_balance')}"
    step("GET /v1/merchant/balance/query (after)", ok, f"{msg} {bal2}".strip())

    print("OpenAPI smoke —", base)
    for ln in lines:
        print(ln)
    if failed:
        print("汇总: FAIL（存在未通过步骤）")
        return 1
    print("汇总: PASS（全部步骤业务 code=2000 且关键接口成功）")
    return 0


def cmd_payin_create(args: argparse.Namespace) -> int:
    mo = args.merchant_order_no or random_merchant_order_no("MO")
    code, text = do_payin_create(
        args.base.rstrip("/"),
        args.app_id,
        args.secret,
        merchant_order_no=mo,
        amount=args.amount,
        currency=args.currency,
        payin_type=args.payin_type,
        notify_url=args.notify_url or "",
        return_url=args.return_url or "",
        subject=args.subject or "",
    )
    print(text)
    ok, _, _, _ = is_api_success(code, text)
    return 0 if ok else 1


def cmd_payout_create(args: argparse.Namespace) -> int:
    mp = args.merchant_order_no or random_merchant_order_no("MP")
    code, text = do_payout_create(
        args.base.rstrip("/"),
        args.app_id,
        args.secret,
        merchant_order_no=mp,
        amount=args.amount,
        currency=args.currency,
        payout_product_code=args.payout_product_code,
        notify_url=args.notify_url or "",
    )
    print(text)
    ok, _, _, _ = is_api_success(code, text)
    return 0 if ok else 1


def cmd_payin_query(args: argparse.Namespace) -> int:
    if not args.order_no and not args.merchant_order_no:
        print("need --order-no or --merchant-order-no", file=sys.stderr)
        return 2
    code, text = do_payin_query(
        args.base.rstrip("/"),
        args.app_id,
        args.secret,
        order_no=args.order_no or "",
        merchant_order_no=args.merchant_order_no or "",
    )
    print(text)
    ok, _, _, _ = is_api_success(code, text)
    return 0 if ok else 1


def cmd_payout_query(args: argparse.Namespace) -> int:
    if not args.order_no and not args.merchant_order_no:
        print("need --order-no or --merchant-order-no", file=sys.stderr)
        return 2
    code, text = do_payout_query(
        args.base.rstrip("/"),
        args.app_id,
        args.secret,
        order_no=args.order_no or "",
        merchant_order_no=args.merchant_order_no or "",
    )
    print(text)
    ok, _, _, _ = is_api_success(code, text)
    return 0 if ok else 1


def cmd_balance(args: argparse.Namespace) -> int:
    code, text = do_balance(args.base.rstrip("/"), args.app_id, args.secret)
    print(text)
    ok, _, _, _ = is_api_success(code, text)
    return 0 if ok else 1


def cmd_check(args: argparse.Namespace) -> int:
    code, text = _get(args.base.rstrip("/") + "/health")
    print(text)
    ok, _, _, _ = is_api_success(code, text)
    return 0 if ok else 1


def _add_common(p: argparse.ArgumentParser) -> None:
    p.add_argument(
        "--base",
        default="http://127.0.0.1:8090",
        help="OpenAPI 网关基址（无尾斜杠），默认 8090",
    )
    p.add_argument("--app-id", default="app_demo", help="商户 app_id（seed: app_demo）")
    p.add_argument("--secret", default="demo_secret", help="api_secret（seed: demo_secret）")


def main() -> int:
    p = argparse.ArgumentParser(description="OpenAPI 签名联调（代收/代付/查单/余额）")
    _add_common(p)

    sub = p.add_subparsers(dest="cmd", required=True)

    s0 = sub.add_parser("run", help="一键跑通：health → 余额 → 代收下单/查单 → 代付下单/查单 → 余额（订单号随机）")
    s0.add_argument("--payin-amount", type=int, default=100, dest="payin_amount")
    s0.add_argument("--payout-amount", type=int, default=100, dest="payout_amount")
    s0.add_argument("--currency", default="CNY")
    s0.add_argument("--payin-type", default="mock", dest="payin_type")
    s0.add_argument("--payout-product-code", default="bank_card", dest="payout_product_code")
    s0.set_defaults(func=cmd_run_all)

    s1 = sub.add_parser("payin-create", help="POST /v1/payin/order")
    s1.add_argument("--merchant-order-no", default="", help="默认随机生成")
    s1.add_argument("--amount", type=int, default=100)
    s1.add_argument("--currency", default="CNY")
    s1.add_argument("--payin-type", default="mock", dest="payin_type")
    s1.add_argument("--notify-url", default="")
    s1.add_argument("--return-url", default="")
    s1.add_argument("--subject", default="")
    s1.set_defaults(func=cmd_payin_create)

    s2 = sub.add_parser("payout-create", help="POST /v1/payout/order")
    s2.add_argument("--merchant-order-no", default="", help="默认随机生成")
    s2.add_argument("--amount", type=int, default=100)
    s2.add_argument("--currency", default="CNY")
    s2.add_argument("--payout-product-code", default="bank_card")
    s2.add_argument("--notify-url", default="")
    s2.set_defaults(func=cmd_payout_create)

    s3 = sub.add_parser("payin-query", help="GET /v1/payin/query")
    s3.add_argument("--order-no", default="")
    s3.add_argument("--merchant-order-no", default="")
    s3.set_defaults(func=cmd_payin_query)

    s4 = sub.add_parser("payout-query", help="GET /v1/payout/query")
    s4.add_argument("--order-no", default="")
    s4.add_argument("--merchant-order-no", default="")
    s4.set_defaults(func=cmd_payout_query)

    s5 = sub.add_parser("balance", help="GET /v1/merchant/balance/query")
    s5.set_defaults(func=cmd_balance)

    s6 = sub.add_parser("check", help="GET /health")
    s6.set_defaults(func=cmd_check)

    args = p.parse_args()
    return args.func(args)


if __name__ == "__main__":
    if len(sys.argv) == 1:
        sys.argv.append("run")
    raise SystemExit(main())
