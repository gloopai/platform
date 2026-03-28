# 部署与联调脚本

## `openapi_smoke.py`

对 **OpenAPI**（gateway `OpenAPIServer`，默认 `http://127.0.0.1:8090`）做联调。签名与 gateway `Md5Sign` 一致；成功以响应 JSON 信封 **`code == 2000`** 为准。

**依赖**：Python 3.10+（仅标准库）。子命令 **`run` / `notify-sim`** 中的上游回调请求发往 **`--base` 同一 OpenAPI 地址**（默认 `http://127.0.0.1:8090`）下的 `POST /v1/callback/upstream/*`。

### 一键全流程（推荐）

无参数即执行 **run**，订单号在脚本内随机生成（`MO-…` / `MP-…`），依次请求：

`GET /health` → `GET /v1/merchant/balance/query` → `POST /v1/payin/order` → `GET /v1/payin/query` → `POST /v1/payout/order` → `GET /v1/payout/query` → `GET /v1/merchant/balance/query`

最后打印每步 **PASS/FAIL** 与汇总 **PASS/FAIL**，退出码 0 表示全部成功。

```bash
python3 openapi_smoke.py
# 等价于
python3 openapi_smoke.py run

# 调整金额与基址
python3 openapi_smoke.py run --base http://127.0.0.1:8090 --payin-amount 200 --payout-amount 200
```

### 单步命令

- `payin-create` / `payout-create`：商户订单号可省略，**默认随机生成**。
- `payin-query` / `payout-query`：需 `--order-no` 或 `--merchant-order-no`。
- `balance`、`check`（health）

```bash
python3 openapi_smoke.py payin-create --amount 100
python3 openapi_smoke.py payout-create --amount 500
```

全局参数：`--base`、`--app-id`、`--secret`（对应 `merchants.app_secret`）。
