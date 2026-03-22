# 聚合支付 · 收银台（Checkout）

面向 **付款用户** 的 H5/PC 收银页：展示订单金额与状态、选择支付方式（占位）、轮询订单结果。与 **商户台 / 管理台** 独立部署，入口一般为下单接口返回的 `checkout_url`。

代码位置：`frontend/apps/checkout`。

---

## 1. 启动与代理

```bash
# 在 frontend 根目录
npm run dev:checkout
```

默认开发服务器见 `frontend/apps/checkout/vite.config.ts`：`/v1` 代理到本机网关 `http://127.0.0.1:8080`，便于联调终端接口。

---

## 2. 路由与参数

| 说明 | 值 |
|------|-----|
| 路由 | `/`（`src/router.ts` 单页） |
| 必填查询参数 | `order_no` — 平台订单号 |

示例：`http://127.0.0.1:<port>/?order_no=xxx`

无 `order_no` 时页面展示友好提示，不请求接口。

---

## 3. 接口

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/v1/terminal/order?order_no=` | 拉取订单展示字段（金额、币种、状态、`return_url`、`merchant_id`、`merchant_order_no`、`pay_product_code` 等），并返回 **`pay_products`**：当前金额下可用的支付产品列表（`code` / `name`），收银台应优先用该列表渲染「选择支付方式」。 |
| POST | `/v1/terminal/pay` | **发起支付（E2）**：Body JSON `order_no`、`pay_product_code`（可选，默认用订单已有编码）。trade 侧按产品重新路由并更新待支付订单的 `channel_id` / `pay_product_*`；响应 `pay_url`、`qr_payload`、`pay_mode`（`mock` / `qr`）。无 `gateway_url` 的通道返回 `mock://…` 载体，前端可生成二维码；有 HTTP `gateway_url` 时拼接 `order_no` 查询参数。 |

前端类型与网关 `OrderInfo` 对齐；改字段时同步 **`gateway` 的 API 定义** 与本文。

---

## 4. 交互与占位说明

- **剩余时间**：当前为进入页面后 **15 分钟** 客户端倒计时（演示用）；与真实订单过期时间对齐时，需后端返回 `expire_at` 等字段后再改前端。
- **待支付轮询**：`status === 0` 且未超时时间隔 2s 刷新订单；文案为「待支付：页面将自动刷新…」，避免与「支付处理中」混淆。
- **超时**：倒计时结束后禁用「唤起支付 / 扫码支付」，并提示「支付已超时」。
- **确认支付**：调用 `POST /v1/terminal/pay`；PC 展示二维码（当前用 `qr_payload` 经公共 QR 图服务生成图片，仅演示）；移动端若 `pay_url` 为 `http(s)` 则整页跳转。真实上游 SDK / 深度链接可在网关侧扩展 `pay_mode`。

---

## 5. 修订记录

| 日期 | 说明 |
|------|------|
| 2026-03-22 | 视觉与信息架构：页头品牌与「加密传输」提示、订单区展示 `merchant_order_no`/`merchant_id`、复制订单号、支付方式单选样式、超时与缺参态、轮询与成功提示文案优化。 |
| 2026-03-22 | 主色与商户台对齐：**slate** 蓝灰（主按钮、Logo、支付方式选中、成功态卡片），整体更稳重。 |
| 2026-03-22 | E2：`POST /v1/terminal/pay` 与收银台发起支付、扫码弹窗展示路由结果。 |

详见仓库 `docs/开发日志.md` 当日条目。
