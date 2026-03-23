# 聚合支付 · 商户平台（Merchant Console）

面向 **接入方商户** 的自助门户：查单、对账、产品与费率说明、账户安全与 **API 联调**。与总管理台（Admin）区分：商户 **仅能访问本商户数据**。

本文与前端 `frontend/apps/merchant` 目录结构对应，便于后续改接口与调试。

---

## 1. 菜单与路由

| 路径 | 说明 | 后端/页面状态 |
|------|------|----------------|
| `/console` | 控制台：今日流水、订单数、成功率、余额 | 已对接 `GET /v1/merchant/summary` |
| `/transactions` | 交易管理：订单列表、详情、回调记录、重发通知 | 已对接 orders / detail / retry_notify |
| `/finance` | 财务中心：资金流水、提现/对账占位 | 已对接 `GET /v1/merchant/fund_logs` |
| `/products` | 产品与费率 | 占位页，文案见 `config/merchantPlaceholder.ts` |
| `/account` | 账户与安全 | 占位页 |
| `/developers` | 开发配置：参数、下单联调、模拟回调、签名工具 | **已对接**开放接口；联调下单可选 **支付产品编码**（`pay_type`），见 §4.1 |

侧栏与底部导航数据来自 **`src/config/merchantMenu.ts`**（唯一数据源），新增菜单请同步修改该文件与 **`src/router.ts`**。

**壳层（`MerchantLayout.vue`）**：桌面端无独立大顶栏，品牌与「退出登录」放在左侧栏；小屏为薄顶栏（Logo + 标题、头像展开账户与退出）加底部 Tab。

**视觉**：以 **slate** 蓝灰为主色（主按钮、品牌块、导航选中态），整体偏稳重商务；与 **收银台** 同一主色体系，仅布局与场景不同。

---

## 2. 前端分层（调试入口）

| 层级 | 路径 | 职责 |
|------|------|------|
| **类型** | `src/types/merchant.api.ts` | 与网关 JSON 字段一致的 TypeScript 类型 |
| **端点常量** | `src/api/endpoints.ts` | `MERCHANT_API.*`、`OPEN_API.*`，改网关路径只改此处 |
| **业务 API** | `src/api/console.ts`、`orders.ts`、`finance.ts`、`session.ts` | 封装 `merchantConsoleGet/Post`，页面不直接拼 URL |
| **HTTP 传输** | `src/lib/http.ts` | Token 请求 + 开放签名请求 |
| **会话** | `src/lib/auth.ts` | localStorage、展示名解析 |
| **工具** | `src/utils/format.ts`、`orderStatus.ts` | 金额、时间、订单状态展示 |
| **组合式** | `src/composables/useMerchantSummary.ts` | 控制台汇总数据加载 |
| **组件** | `components/layout/PageHeader.vue`、`components/ui/ErrorCallout.vue` | 统一页头与错误提示 |

别名 **`@/`** 指向 `src/`，见 `vite.config.ts` 与 `tsconfig.app.json`。

订单相关接口中的 **`pay_product_code`** 表示对外「支付产品」编码（与开放 API `pay_type` 一致），与内部 `channel_id`（上游实例）不同，见 [`通道与支付产品.md`](./通道与支付产品.md)。

---

## 3. 接口一览（商户控制台 Token）

使用请求头 `X-Merchant-Token`（由 `lib/http.ts` 注入）：

- `GET /v1/merchant/summary`
- `GET /v1/merchant/orders`
- `GET /v1/merchant/order/detail`
- `POST /v1/merchant/order/retry_notify`
- `GET /v1/merchant/fund_logs`
- `POST /v1/merchant/logout`

登录为 **`POST /v1/merchant/login`**（无 Token，见 `LoginPage.vue`）。

---

## 4. 开放接口（联调，非 Token）

在 **`api/endpoints.ts`** 的 `OPEN_API` 中维护：

- `POST /v1/pay/order` — 下单
- `POST /v1/callback/notify` — 模拟上游回调（开发页）

`DevelopersPage.vue` 中已改为引用 `OPEN_API`，避免硬编码散落。

### 4.1 下单与「支付产品」编码（阶段 D1 / D2）

- 创建订单请求体中的 **`pay_type` 表示支付产品编码**（与 `pay_products.code`、收银台可选列表一致），**不是**内部上游实例 ID；商户开放 API **不要求**传 `channel_id` 参与选路（路由在平台侧完成）。
- 演示用可选值集中在 **`src/config/payProducts.ts`**（`DEMO_PAY_PRODUCT_OPTIONS`），应与演示库 `seed_demo.sql` 中的产品行保持同步；需要其他编码时使用联调页「自定义编码」。
- 模拟回调区块中的 **`channel_id`** 仅用于**扮演上游**通知平台，与「商户下单不传 channel」不矛盾。

---

## 5. 占位模块扩展方式

1. 在 `config/merchantPlaceholder.ts` 增加 `路径 → 文案`。  
2. 在 `router.ts` 增加子路由，`component: ModulePlaceholderPage`。  
3. 在 `config/merchantMenu.ts` 增加导航项与 `merchantPathTitle`。  

待后端就绪后，将占位路由替换为真实页面组件，并新增 `src/api/xxx.ts` 中的请求函数。

---

## 6. 修订说明

代码组织以「**类型集中、API 按域拆分、页面只调 api 层**」为原则；修改网关路径时优先检查 **`api/endpoints.ts`** 与 **`types/merchant.api.ts`**。
