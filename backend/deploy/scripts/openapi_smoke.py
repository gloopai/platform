#!/usr/bin/env python3
"""
OpenAPI 联调脚本：代收/代付下单与查单、余额（gateway OpenAPIServer，默认 :8090）。

- 签名与 gateway internal/middleware/sign_md5.go 的 Md5Sign 一致。
- 代收/代付创建接口不允许携带 channel_id（上游路由由平台决定）；请勿在请求体中包含该字段。
- 业务成功以统一 JSON 信封 code=2000 为准（与 apiresp.CodeSuccess 一致）。

主入口（默认 `run` = health/余额 + 代收 3 笔 + 代付 3 笔 + 向同一 OpenAPI 基址投递 `/v1/callback/upstream/*` mock 上游通知，覆盖成功/处理中/失败）：

    python3 openapi_smoke.py
    python3 openapi_smoke.py run

仅跑上游通知场景（无余额前后对比）：

    python3 openapi_smoke.py notify-sim

需可访问 OpenAPI（默认 8090；`/v1/callback/*` 与验签接口同端口）。seed_demo 仅保留 mock-psp-alt：代收/代付上游回调均为 mock_psp_alt 的 MD5+snake_case；--channel-id 与 --payout-channel-id 一般同为 `SELECT id FROM channels WHERE name='mock-psp-alt'`（空库常为 1，删过旧 mock-psp 后常为 2）。

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
import urllib.parse
from typing import Any

API_CODE_SUCCESS = 2000


def md5_sign(params: dict[str, str], secret: str) -> str:
    keys = sorted(k.lower() for k in params if k.lower() not in ("sign", "signature"))
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


def _payin_notify_state_word(status: str) -> str:
    """mock_psp_alt 代收异步 state：PENDING/SUCCESS/FAIL；兼容旧脚本 1/2/3。"""
    s = status.strip().upper()
    if s in ("1", "PROCESSING", "PENDING"):
        return "PENDING"
    if s in ("2", "SUCCESS"):
        return "SUCCESS"
    if s in ("3", "FAIL", "FAILED"):
        return "FAIL"
    return s


def build_mock_payin_notify(
    channel_secret: str,
    *,
    platform_order_no: str,
    sys_order_no: str,
    status: str,
    amount_minor: int,
) -> dict[str, Any]:
    """
    mock_psp_alt（mockpsp2）代收异步：snake_case + MD5（与 channeldriver/mockpsp2.VerifyPayinNotify 一致）。
    若通道仍为 mock_psp（mock-psp），需改用 camelCase+HMAC，勿与此混用。
    """
    ts = str(int(time.time() * 1000))
    amt = str(amount_minor)
    st = _payin_notify_state_word(status)
    sign_params: dict[str, str] = {
        "amount": amt,
        "event_time": ts,
        "merchant_ref": platform_order_no,
        "state": st,
        "txn_id": sys_order_no,
    }
    sig = md5_sign(sign_params, channel_secret)
    return {
        "merchant_ref": platform_order_no,
        "txn_id": sys_order_no,
        "state": st,
        "amount": amt,
        "event_time": ts,
        "signature": sig,
    }


def _payout_notify_state_word(status: str) -> str:
    """mock_psp_alt（mockpsp2）用 PROCESSING/SUCCESS/FAIL；兼容旧脚本 1/2/3。"""
    s = status.strip().upper()
    if s in ("1", "PROCESSING"):
        return "PROCESSING"
    if s in ("2", "SUCCESS"):
        return "SUCCESS"
    if s in ("3", "FAIL", "FAILED"):
        return "FAIL"
    return s


def build_mock_payout_notify(
    channel_secret: str,
    *,
    platform_order_no: str,
    sys_order_no: str,
    status: str,
    amount_minor: int,
    reference_no: str,
) -> dict[str, Any]:
    """
    mock_psp_alt 代付异步：snake_case + MD5（与 channeldriver/mockpsp2.VerifyPayoutNotify 一致）。
    mock_psp（仅 mock-psp）为 camelCase+HMAC，勿混用。
    """
    ts = str(int(time.time() * 1000))
    amt = str(amount_minor)
    pst = _payout_notify_state_word(status)
    br = reference_no.strip()
    sign_params: dict[str, str] = {
        "amount": amt,
        "bank_reference": br,
        "event_time": ts,
        "merchant_ref": platform_order_no,
        "payout_state": pst,
        "txn_id": sys_order_no,
    }
    sig = md5_sign(sign_params, channel_secret)
    return {
        "merchant_ref": platform_order_no,
        "txn_id": sys_order_no,
        "payout_state": pst,
        "amount": amt,
        "bank_reference": reference_no,
        "event_time": ts,
        "signature": sig,
    }


def upstream_callback_url(openapi_base: str, kind: str, channel_id: int, platform_order_no: str) -> str:
    on = urllib.parse.quote(platform_order_no, safe="")
    path = "/v1/callback/upstream/payin" if kind == "payin" else "/v1/callback/upstream/payout"
    return f"{openapi_base.rstrip('/')}{path}?channel_id={channel_id}&order_no={on}"


def post_upstream_json(openapi_base: str, kind: str, channel_id: int, platform_order_no: str, body: dict[str, Any]) -> tuple[int, str]:
    url = upstream_callback_url(openapi_base, kind, channel_id, platform_order_no)
    return _post_json(url, body)


def query_order_status(base: str, app_id: str, secret: str, *, payin: bool, order_no: str) -> int | None:
    if payin:
        hc, ht = do_payin_query(base, app_id, secret, order_no=order_no)
    else:
        hc, ht = do_payout_query(base, app_id, secret, order_no=order_no)
    ok, _, _, data = is_api_success(hc, ht)
    if not ok:
        return None
    o = data.get("order")
    if not isinstance(o, dict):
        return None
    st = o.get("status")
    return int(st) if isinstance(st, int) else None


def _with_sign(
    base: dict[str, str],
    *,
    app_id: str,
    app_secret: str,
) -> dict[str, str]:
    p = {k.lower(): str(v).strip() for k, v in base.items() if v is not None}
    p["app_id"] = app_id
    p["timestamp"] = str(int(time.time()))
    p["nonce"] = secrets.token_hex(8)
    p["sign"] = md5_sign(p, app_secret)
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


def _http_request(
    method: str,
    url: str,
    *,
    body: bytes | None = None,
    timeout: float = 30.0,
) -> tuple[int, str]:
    """用 http.client 发请求。Python 3.13+ 上 urllib.request.urlopen 对本地 HTTP 常误报 RemoteDisconnected。"""
    parsed = urllib.parse.urlparse(url)
    scheme = (parsed.scheme or "http").lower()
    if scheme not in ("http", "https"):
        sys.stderr.write(f"不支持的 URL scheme: {scheme!r}（仅支持 http/https）\n")
        raise SystemExit(1)
    host = parsed.hostname
    if not host:
        sys.stderr.write(f"非法 URL: {url}\n")
        raise SystemExit(1)
    port = parsed.port
    if port is None:
        port = 443 if scheme == "https" else 80
    path = parsed.path or "/"
    if parsed.query:
        path += "?" + parsed.query

    headers: dict[str, str] = {"Accept": "application/json", "Connection": "close"}
    if body is not None:
        headers["Content-Type"] = "application/json"

    Conn = http.client.HTTPSConnection if scheme == "https" else http.client.HTTPConnection
    try:
        conn = Conn(host, port, timeout=timeout)
        try:
            conn.request(method, path, body=body, headers=headers)
            resp = conn.getresponse()
            text = resp.read().decode("utf-8")
            return resp.status, text
        finally:
            conn.close()
    except (http.client.RemoteDisconnected, ConnectionResetError, BrokenPipeError, TimeoutError, OSError) as e:
        sys.stderr.write(_conn_error_hint(url, e))
        raise SystemExit(1) from None


def _post_json(url: str, body: dict[str, Any], timeout: float = 30.0) -> tuple[int, str]:
    data = json.dumps(body, ensure_ascii=False).encode("utf-8")
    return _http_request("POST", url, body=data, timeout=timeout)


def _get(url: str, timeout: float = 30.0) -> tuple[int, str]:
    return _http_request("GET", url, timeout=timeout)


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


def api_error_hint(http_code: int, raw: str, ok: bool, code: int | None, msg: str) -> str:
    """业务失败时附加说明（常见：4223 无可用通道需跑 seed、4013 验签、4091 重放、5003 Redis）。"""
    if ok:
        return ""
    if http_code != 200:
        return f" | {raw[:500]!r}" if raw.strip() else f" | HTTP {http_code} 空 body"
    return f" | code={code} msg={msg!r}"


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
    p = _with_sign(params, app_id=app_id, app_secret=secret)
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
    flat = {k: v for k, v in _with_sign(b, app_id=app_id, app_secret=secret).items() if v != ""}
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
    flat = {k: v for k, v in _with_sign(b, app_id=app_id, app_secret=secret).items() if v != ""}
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


def _append_step(lines: list[str], failed: list[bool], name: str, ok: bool, detail: str = "") -> None:
    """failed 为单元素列表，便于在嵌套函数中修改。"""
    if not ok:
        failed[0] = True
    mark = "PASS" if ok else "FAIL"
    line = f"  [{mark}] {name}"
    if detail:
        line += f" — {detail}"
    lines.append(line)


def run_notify_scenario_steps(args: argparse.Namespace) -> tuple[bool, list[str]]:
    """
    代收 3 笔 + 代付 3 笔，分别投递上游异步通知（mock_psp_alt / MD5），覆盖成功/处理中/失败。
    代收回调：channel_id + channel_sign_secret；代付：payout_channel_id + payout_channel_sign_secret。
    seed 仅 mock-psp-alt 时两组参数常相同（见 --channel-id / --payout-channel-id）。
    """
    base = args.base.rstrip("/")
    cid = args.channel_id
    chsec = args.channel_sign_secret
    payout_cid = getattr(args, "payout_channel_id", cid)
    payout_chsec = getattr(args, "payout_channel_sign_secret", chsec)
    app_id = args.app_id
    secret = args.secret
    payin_amt = args.payin_amount
    payout_amt = args.payout_amount
    lines: list[str] = []
    failed = [False]

    def step(name: str, ok: bool, detail: str = "") -> None:
        _append_step(lines, failed, name, ok, detail)

    # --- Payin x3 ---
    # 1) success → 平台订单应变更为已支付
    mo_ok = random_merchant_order_no("MO-NP")
    pc, pt = do_payin_create(
        base,
        app_id,
        secret,
        merchant_order_no=mo_ok,
        amount=payin_amt,
        currency=args.currency,
        payin_type=args.payin_type,
    )
    ok_c, code_c, msg_c, pdata = is_api_success(pc, pt)
    pno_ok = str(pdata.get("order_no") or "") if pdata else ""
    step(
        "POST payin (success case)",
        ok_c and bool(pno_ok),
        f"merchant_order_no={mo_ok} order_no={pno_ok}{api_error_hint(pc, pt, ok_c, code_c, msg_c)}",
    )
    if ok_c and pno_ok:
        body = build_mock_payin_notify(
            chsec,
            platform_order_no=pno_ok,
            sys_order_no=f"MOCK-NP-OK-{pno_ok}",
            status="2",
            amount_minor=payin_amt,
        )
        uhc, uht = post_upstream_json(base, "payin", cid, pno_ok, body)
        ub_ok = uhc == 200 and uht.strip() == "SUCCESS"
        step("upstream payin notify status=2 (success)", ub_ok, f"HTTP {uhc} body={uht.strip()!r}")
        st = query_order_status(base, app_id, secret, payin=True, order_no=pno_ok)
        step("payin query status==1 (paid)", st == 1, f"status={st!r}")

    # 2) processing → 网关拒单（非成功态不落账），订单仍待支付
    mo_pr = random_merchant_order_no("MO-NP")
    pc2, pt2 = do_payin_create(
        base,
        app_id,
        secret,
        merchant_order_no=mo_pr,
        amount=payin_amt,
        currency=args.currency,
        payin_type=args.payin_type,
    )
    ok_c2, code_c2, msg_c2, pd2 = is_api_success(pc2, pt2)
    pno_pr = str(pd2.get("order_no") or "") if pd2 else ""
    step(
        "POST payin (processing case)",
        ok_c2 and bool(pno_pr),
        f"order_no={pno_pr}{api_error_hint(pc2, pt2, ok_c2, code_c2, msg_c2)}",
    )
    if ok_c2 and pno_pr:
        body2 = build_mock_payin_notify(
            chsec,
            platform_order_no=pno_pr,
            sys_order_no=f"MOCK-NP-PR-{pno_pr}",
            status="1",
            amount_minor=payin_amt,
        )
        uhc2, uht2 = post_upstream_json(base, "payin", cid, pno_pr, body2)
        ub2 = uhc2 == 200 and uht2.strip() == "FAIL"
        step("upstream payin notify status=1 (processing → PSP FAIL)", ub2, f"HTTP {uhc2} body={uht2.strip()!r}")
        st2 = query_order_status(base, app_id, secret, payin=True, order_no=pno_pr)
        step("payin query still pending (0)", st2 == 0, f"status={st2!r}")

    # 3) failed notify → 不落账
    mo_f = random_merchant_order_no("MO-NP")
    pc3, pt3 = do_payin_create(
        base,
        app_id,
        secret,
        merchant_order_no=mo_f,
        amount=payin_amt,
        currency=args.currency,
        payin_type=args.payin_type,
    )
    ok_c3, code_c3, msg_c3, pd3 = is_api_success(pc3, pt3)
    pno_f = str(pd3.get("order_no") or "") if pd3 else ""
    step(
        "POST payin (failed-notify case)",
        ok_c3 and bool(pno_f),
        f"order_no={pno_f}{api_error_hint(pc3, pt3, ok_c3, code_c3, msg_c3)}",
    )
    if ok_c3 and pno_f:
        body3 = build_mock_payin_notify(
            chsec,
            platform_order_no=pno_f,
            sys_order_no=f"MOCK-NP-FL-{pno_f}",
            status="3",
            amount_minor=payin_amt,
        )
        uhc3, uht3 = post_upstream_json(base, "payin", cid, pno_f, body3)
        ub3 = uhc3 == 200 and uht3.strip() == "FAIL"
        step("upstream payin notify status=3 (failed → PSP FAIL)", ub3, f"HTTP {uhc3} body={uht3.strip()!r}")
        st3 = query_order_status(base, app_id, secret, payin=True, order_no=pno_f)
        step("payin query still pending (0)", st3 == 0, f"status={st3!r}")

    # --- Payout x3 ---
    # 4) success
    mp_ok = random_merchant_order_no("MP-NP")
    poc, pot = do_payout_create(
        base,
        app_id,
        secret,
        merchant_order_no=mp_ok,
        amount=payout_amt,
        currency=args.currency,
        payout_product_code=args.payout_product_code,
    )
    pok, code_po, msg_po, pod = is_api_success(poc, pot)
    po_no = str(pod.get("order_no") or "") if pod else ""
    step(
        "POST payout (success case)",
        pok and bool(po_no),
        f"merchant_order_no={mp_ok} order_no={po_no}{api_error_hint(poc, pot, pok, code_po, msg_po)}",
    )
    if pok and po_no:
        pb = build_mock_payout_notify(
            payout_chsec,
            platform_order_no=po_no,
            sys_order_no=f"MOCK-NP-POK-{po_no}",
            status="2",
            amount_minor=payout_amt,
            reference_no="UTR-OK-001",
        )
        phc, pht = post_upstream_json(base, "payout", payout_cid, po_no, pb)
        pb_ok = phc == 200 and pht.strip() == "SUCCESS"
        step("upstream payout notify status=2 (success)", pb_ok, f"HTTP {phc} body={pht.strip()!r}")
        pst = query_order_status(base, app_id, secret, payin=False, order_no=po_no)
        step("payout query status==1 (success)", pst == 1, f"status={pst!r}")

    # 5) processing（仅 ACK，库表仍为处理中）
    mp_pr = random_merchant_order_no("MP-NP")
    poc2, pot2 = do_payout_create(
        base,
        app_id,
        secret,
        merchant_order_no=mp_pr,
        amount=payout_amt,
        currency=args.currency,
        payout_product_code=args.payout_product_code,
    )
    pok2, code_po2, msg_po2, pod2 = is_api_success(poc2, pot2)
    po_pr = str(pod2.get("order_no") or "") if pod2 else ""
    step(
        "POST payout (processing case)",
        pok2 and bool(po_pr),
        f"order_no={po_pr}{api_error_hint(poc2, pot2, pok2, code_po2, msg_po2)}",
    )
    if pok2 and po_pr:
        pb2 = build_mock_payout_notify(
            payout_chsec,
            platform_order_no=po_pr,
            sys_order_no=f"MOCK-NP-PPR-{po_pr}",
            status="1",
            amount_minor=payout_amt,
            reference_no="",
        )
        phc2, pht2 = post_upstream_json(base, "payout", payout_cid, po_pr, pb2)
        pb2_ok = phc2 == 200 and pht2.strip() == "SUCCESS"
        step("upstream payout notify status=1 (processing ACK)", pb2_ok, f"HTTP {phc2} body={pht2.strip()!r}")
        pst2 = query_order_status(base, app_id, secret, payin=False, order_no=po_pr)
        step("payout query still pending (0)", pst2 == 0, f"status={pst2!r}")

    # 6) failed
    mp_fl = random_merchant_order_no("MP-NP")
    poc3, pot3 = do_payout_create(
        base,
        app_id,
        secret,
        merchant_order_no=mp_fl,
        amount=payout_amt,
        currency=args.currency,
        payout_product_code=args.payout_product_code,
    )
    pok3, code_po3, msg_po3, pod3 = is_api_success(poc3, pot3)
    po_fl = str(pod3.get("order_no") or "") if pod3 else ""
    step(
        "POST payout (failed case)",
        pok3 and bool(po_fl),
        f"order_no={po_fl}{api_error_hint(poc3, pot3, pok3, code_po3, msg_po3)}",
    )
    if pok3 and po_fl:
        pb3 = build_mock_payout_notify(
            payout_chsec,
            platform_order_no=po_fl,
            sys_order_no=f"MOCK-NP-PFL-{po_fl}",
            status="3",
            amount_minor=payout_amt,
            reference_no="",
        )
        phc3, pht3 = post_upstream_json(base, "payout", payout_cid, po_fl, pb3)
        pb3_ok = phc3 == 200 and pht3.strip() == "SUCCESS"
        step("upstream payout notify status=3 (failed)", pb3_ok, f"HTTP {phc3} body={pht3.strip()!r}")
        pst3 = query_order_status(base, app_id, secret, payin=False, order_no=po_fl)
        step("payout query status==2 (failed)", pst3 == 2, f"status={pst3!r}")

    return failed[0], lines


def cmd_notify_sim(args: argparse.Namespace) -> int:
    """仅跑上游通知多场景（另含 health）。"""
    base = args.base.rstrip("/")
    cid = args.channel_id
    lines: list[str] = []
    failed = [False]

    def step(name: str, ok: bool, detail: str = "") -> None:
        _append_step(lines, failed, name, ok, detail)

    hc, ht = _get(base + "/health")
    ok, code, msg, _ = is_api_success(hc, ht)
    step("GET /health (openapi)", ok, f"code={code} {msg}" if code is not None else ht[:120])

    nf, slines = run_notify_scenario_steps(args)
    lines.extend(slines)
    failed[0] = failed[0] or nf

    print(
        "OpenAPI + upstream notify simulation —",
        base,
        "| payin_ch",
        cid,
        "| payout_ch",
        getattr(args, "payout_channel_id", cid),
    )
    for ln in lines:
        print(ln)
    if failed[0]:
        print("汇总: FAIL")
        return 1
    print("汇总: PASS（代收/代付多笔 + 上游回调与查单状态符合预期）")
    return 0


def cmd_run_all(args: argparse.Namespace) -> int:
    """默认一键：health → 余额 → 代收/代付各 3 笔 + 上游回调模拟全部状态 → 余额。"""
    base = args.base.rstrip("/")
    cid = args.channel_id
    app_id = args.app_id
    secret = args.secret
    lines: list[str] = []
    failed = [False]

    def step(name: str, ok: bool, detail: str = "") -> None:
        _append_step(lines, failed, name, ok, detail)

    hc, ht = _get(base + "/health")
    ok, code, msg, _ = is_api_success(hc, ht)
    step("GET /health", ok, f"code={code} {msg}" if code is not None else ht[:200])

    bc, bt = do_balance(base, app_id, secret)
    ok, code, msg, bdata = is_api_success(bc, bt)
    bal_detail = ""
    if ok and bdata:
        bal_detail = f"available={bdata.get('available_balance')} payin={bdata.get('payin_balance')}"
    step("GET /v1/merchant/balance/query (before)", ok, f"{msg} {bal_detail}".strip())

    nf, slines = run_notify_scenario_steps(args)
    lines.extend(slines)
    failed[0] = failed[0] or nf

    ac, at = do_balance(base, app_id, secret)
    ok, code, msg, adata = is_api_success(ac, at)
    bal2 = ""
    if ok and adata:
        bal2 = f"available={adata.get('available_balance')}"
    step("GET /v1/merchant/balance/query (after)", ok, f"{msg} {bal2}".strip())

    print(
        "OpenAPI smoke —",
        base,
        "| payin_ch",
        cid,
        "| payout_ch",
        getattr(args, "payout_channel_id", cid),
    )
    for ln in lines:
        print(ln)
    if failed[0]:
        print("汇总: FAIL（存在未通过步骤）")
        return 1
    print("汇总: PASS（health/余额 + 代收代付各 3 笔状态模拟 + 上游回调）")
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
    p.add_argument("--secret", default="demo_secret", help="app_secret（merchants.app_secret，seed: demo_secret）")


def main() -> int:
    p = argparse.ArgumentParser(description="OpenAPI 签名联调（代收/代付/查单/余额）")
    _add_common(p)

    sub = p.add_subparsers(dest="cmd", required=True)

    s0 = sub.add_parser(
        "run",
        help="一键跑通：health → 余额 → 代收/代付各 3 笔 + mock 上游回调（全部状态）→ 余额",
    )
    s0.add_argument(
        "--channel-id",
        type=int,
        default=1,
        help="代收上游回调 ?channel_id=（mock-psp-alt 的 id；删过旧 mock-psp 后常为 2）",
    )
    s0.add_argument(
        "--channel-sign-secret",
        default="channel_secret_alt",
        help="mock-psp-alt 的 sign_secret（与 MD5 验签一致，seed: channel_secret_alt）",
    )
    s0.add_argument(
        "--payout-channel-id",
        type=int,
        default=1,
        dest="payout_channel_id",
        help="代付上游回调 ?channel_id=（通常与 --channel-id 相同）",
    )
    s0.add_argument(
        "--payout-channel-sign-secret",
        default="channel_secret_alt",
        dest="payout_channel_sign_secret",
        help="代付 mock-psp-alt 的 sign_secret（seed: channel_secret_alt）",
    )
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

    s7 = sub.add_parser(
        "notify-sim",
        help="多笔代收/代付 + 向同一 --base(OpenAPI) 投递 /v1/callback/upstream/* mock 上游回调（MD5），覆盖成功/处理中/失败",
    )
    s7.add_argument(
        "--channel-id",
        type=int,
        default=1,
        help="代收：mock-psp-alt 的 channels.id（不对则查库）",
    )
    s7.add_argument(
        "--channel-sign-secret",
        default="channel_secret_alt",
        help="mock-psp-alt sign_secret（seed: channel_secret_alt）",
    )
    s7.add_argument(
        "--payout-channel-id",
        type=int,
        default=1,
        dest="payout_channel_id",
        help="代付：通常与 --channel-id 相同",
    )
    s7.add_argument(
        "--payout-channel-sign-secret",
        default="channel_secret_alt",
        dest="payout_channel_sign_secret",
        help="代付：mock-psp-alt 的 sign_secret（seed: channel_secret_alt）",
    )
    s7.add_argument("--payin-amount", type=int, default=100, dest="payin_amount")
    s7.add_argument("--payout-amount", type=int, default=100, dest="payout_amount")
    s7.add_argument("--currency", default="CNY")
    s7.add_argument("--payin-type", default="mock", dest="payin_type")
    s7.add_argument("--payout-product-code", default="bank_card", dest="payout_product_code")
    s7.set_defaults(func=cmd_notify_sim)

    args = p.parse_args()
    return args.func(args)


if __name__ == "__main__":
    if len(sys.argv) == 1:
        sys.argv.append("run")
    raise SystemExit(main())
