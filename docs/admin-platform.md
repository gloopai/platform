# 聚合支付 · 总管理后台（Admin Console）

面向平台运营与风控人员的 **总管理台**，用于聚合多家 **上游支付通道**，向 **商户** 提供统一接入、路由、对账与结算能力。本文描述 **功能结构、菜单与路由、与后端对接状态**，与前端 `frontend/apps/admin/src/adminMenu.ts` 保持一致。

---

## 1. 产品定位

| 维度 | 说明 |
|------|------|
| 使用者 | 平台内部运营、财务、技术、风控、客服（按 RBAC 区分） |
| 与商户端 | 商户端仅见本商户数据；总后台可 **跨商户** 检索与操作 |
| 与通道 | 维护通道实例、费率、限额、健康状态；**路由策略** 决定订单如何分配通道 |
| 聚合含义 | 同一商户可配置多种支付方式；平台侧统一订单、清结算与风控 |

---

## 2. 菜单树（路由）

> 路径均为前端 `history` 根路径下的 **绝对路径**。

### 2.1 系统概览

| 路径 | 页面 | 说明 |
|------|------|------|
| `/stats` | 系统概览 | **已对接**：`GET /v1/admin/stats/overview` 今日订单汇总、按支付产品与按上游通道拆解、成交率与支付成功率；启用/熔断通道数 |

### 2.2 商户与接入

| 路径 | 页面 | 说明 |
|------|------|------|
| `/merchants` | 商户管理 | **已对接**：网关 `/v1/admin/merchants` 列表/创建/更新 |
| `/merchant-products` | 代收产品与上游通道 | **已对接**：`pay_products` CRUD 与 `pay_product_channels` 绑定（权重、启用）；见 §4 |
| `/merchant-payout-products` | 代付产品与上游通道 | **已对接**：`payout_products` 及绑定 API；与代收共用 `PayProductsPage`（`payoutMode`） |

### 2.3 通道与路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/channels` | 通道管理 | **已对接**：`/v1/admin/channels` 列表/创建/更新 |
| `/routing` | 路由策略 | **已对接**：说明当前「产品内加权、商户白名单、通道熔断」的实现方式；`GET /v1/admin/routing/summary` 汇总表数据；规则引擎类能力见页面内「后续规划」 |
| `/channel-health` | 通道监控 | **已对接（MVP）**：`GET /v1/admin/routing/summary` + `GET /v1/admin/channels` 只读汇总与通道启用/熔断列；无时序指标 |

### 2.4 交易与资金

| 路径 | 页面 | 说明 |
|------|------|------|
| `/pay-orders` | Pay订单 | **已对接（MVP）**：`GET /v1/admin/pay_orders` 跨商户列表；关键词与商户 ID、状态筛选；只读无导出 |
| `/refunds` | 退款与差错 | **已对接（MVP）**：`GET /v1/admin/refunds` 候选订单只读列表（失败/关闭），支持商户与关键词筛选 |
| `/reconcile` | 对账中心 | **已对接（MVP）**：`GET /v1/admin/reconcile/day?date=YYYY-MM-DD&merchant_id=` 按自然日平台侧订单聚合（可选按商户过滤，与系统概览同源）；上游文件导入与差异批次为后续 |
| `/settlement` | 结算与提现 | **已对接（MVP）**：`GET /v1/admin/settlement/logs?merchant_id=&limit=` 平台资金流水只读（可按商户筛选） |

> **说明**：侧栏暂不挂「风控与合规」等占位菜单，优先跑通收单—订单—对账主路径；风控、审计、公告等能力见 §3.4 规划。

### 2.5 系统与运维

| 路径 | 页面 | 说明 |
|------|------|------|
| `/system` | 系统管理 | **已对接（MVP）**：`GET /v1/admin/admin_users` 管理员账号只读列表（无密码）；RBAC/审计为后续 |
| `/ops` | 运维监控 | **已对接（MVP）**：`GET /health` 网关探活 JSON；无 QPS/链路 |

---

## 3. 功能模块（按市场常见聚合支付）

### 3.1 商户与接入

- **商户主数据**：商户号、密钥、状态、费率、结算周期。
- **产品与签约**：可见支付方式（微信/支付宝/银联等）、产品包、签约/实名状态（若对接进件）。
- **API 与安全**：IP 白名单、回调 URL、密钥轮换。

### 3.2 通道与路由

- **通道管理**：上游通道编码、密钥、限额、权重、费率成本、启用/熔断。
- **路由策略**：规则引擎（金额分段、时间窗、商户等级、A/B）、主备与 failover。
- **监控**：成功率、RT、错误码分布、熔断状态。

### 3.3 交易与资金

- **订单**：全站查询、补单、关单、人工置成功（强权限）。
- **退款与差错**：退款单、撤销、差错与调账。
- **对账**：通道对账文件、平台账、差异类型与处理闭环。
- **结算与提现**：结算单、提现审核、打款通道、手续费。

### 3.4 风控与合规

- **规则**：单笔/单日限额、频次、黑名单、设备指纹、反洗钱报送（视监管要求）。
- **审计**：操作人、时间、对象、前后快照、不可篡改存储（规划）。

### 3.5 系统与运维

- **系统管理**：RBAC、数据权限（按商户）、参数中心。
- **运维**：网关与 RPC 监控、日志、链路追踪、容量告警。

---

## 4. 当前后端对接情况（Gateway）

已实现（示例路径，以仓库 `routes.go` 为准）：

- `POST /v1/admin/login`、`POST /v1/admin/logout`
- `GET/POST /v1/admin/channels`、`PUT /v1/admin/channels/:id`
- `GET/POST /v1/admin/merchants`、`PUT /v1/admin/merchants/:merchant_id`
- **支付产品与通道绑定**（对外 `code` 落库 `pay_products`，绑定表 `pay_product_channels`）：
  - `GET /v1/admin/pay_products`、`POST /v1/admin/pay_products`、`PUT /v1/admin/pay_products/:id`
  - `GET /v1/admin/pay_products/:id/bindings`、`POST /v1/admin/pay_products/:id/bindings`（同 `(product, channel)` 唯一则更新权重/启用）
  - `PUT /v1/admin/pay_product_bindings/:id`、`DELETE /v1/admin/pay_product_bindings/:id`
- **路由策略概览**：`GET /v1/admin/routing/summary`（当前算法标识、各表计数，供「路由策略」页展示）
- **系统概览统计**：`GET /v1/admin/stats/overview`（今日 `pay_orders` 聚合：总额、笔数、状态分布；按 `pay_product_code`、按 `channel_id` 分组）
- **Pay订单（只读）**：`GET /v1/admin/pay_orders?keyword=&merchant_id=&status=&limit=`（`status` 省略为不限状态；trade `AdminListPayOrders`）
- **退款候选（只读）**：`GET /v1/admin/refunds?merchant_id=&keyword=&status=&limit=`（失败/关闭订单初筛，MVP）
- **对账（平台账按日）**：`GET /v1/admin/reconcile/day?date=YYYY-MM-DD&merchant_id=`（trade `Order.AdminDayOverview`，可选商户过滤；与 `stats/overview` 同口径聚合，可选历史自然日）
- **探活（无需管理 Token）**：`GET /health`（JSON：`status`、`service`、`timestamp_ms`；供运维与「运维监控」页）
- **管理员账号（只读）**：`GET /v1/admin/admin_users`（`id`、`username`、`status`；不含密码哈希）
- **结算流水（只读）**：`GET /v1/admin/settlement/logs?merchant_id=&limit=`（`fund_logs` 聚合；MVP 不含结算单与提现流）

其余未列「已对接」的能力见各页内「后续规划」说明。通用占位组件 `ModulePlaceholderPage` 仍保留在仓库供后续模块使用。

---

## 5. 前端实现说明

开发约定（鉴权请求封装、按模块分目录与组件拆分）见 [**管理端前端开发规范**](./管理端前端开发规范.md)。

| 文件 | 作用 |
|------|------|
| `src/adminMenu.ts` | 侧栏菜单结构、面包屑标题、占位页文案 |
| `src/views/AdminLayout.vue` | 多级侧栏、折叠态悬浮子菜单 |
| `src/views/modules/pay-products/` | 支付产品与上游通道绑定（`PayProductsPage.vue` + 子组件） |
| `src/views/modules/channels/` | 通道管理（`ChannelsPage.vue` + `ChannelList` / `ChannelFormCard` 等） |
| `src/views/modules/routing/` | 路由策略说明页（`RouteStrategyPage.vue` + 概览统计与配置入口卡片） |
| `src/views/modules/stats/` | 系统概览（`StatsPage.vue` + KPI / 状态条 / 产品·通道双表） |
| `src/views/modules/orders/` | Pay订单（`OrdersPage.vue` + `types.ts`） |
| `src/views/modules/channel-health/` | 通道监控（`ChannelHealthPage.vue`，复用 `routing/RoutingStatGrid`） |
| `src/views/modules/refunds/` | 退款与差错（`RefundsPage.vue`，候选订单表 + 筛选） |
| `src/views/modules/reconcile/` | 对账中心（`ReconcilePage.vue`，按日平台账（可选商户过滤）+ 复用 `stats` 拆解表） |
| `src/views/modules/settlement/` | 结算与提现（`SettlementPage.vue`，资金流水表 + 商户筛选） |
| `src/views/modules/ops/` | 运维监控（`OpsPage.vue`，`GET /health`） |
| `src/views/modules/system/` | 系统管理（`SystemPage.vue`，`GET /v1/admin/admin_users`） |
| `src/views/pages/ModulePlaceholderPage.vue` | 通用占位页（可选读 `adminPlaceholderMeta`） |
| `src/router.ts` | 路由注册 |

---

## 6. 迭代建议（Roadmap）

1. **数据与报表**：总览大盘对接 `/v1/admin/dashboard` 或直连统计服务。
2. **订单全站**：`pay_orders` 与 trade 服务查询接口，支持导出与权限。
3. **路由与通道**：路由规则存储与执行引擎；通道健康时序指标。
4. **对账与结算**：对账任务表、文件存储、结算单状态机。
5. **RBAC 与审计**：管理员表、角色、菜单与操作日志表。

---

## 7. 修订记录

- **2026-03-23**：`/refunds` 对接 `GET /v1/admin/refunds`（失败/关闭候选订单只读）。
- **2026-03-23**：`/settlement` 对接 `GET /v1/admin/settlement/logs`（平台资金流水只读）。
- **2026-03-23**：`/reconcile` 支持 `merchant_id` 可选过滤（按商户核对平台账）。
- **2026-03-23**：`/system` 系统管理 MVP：`GET /v1/admin/admin_users`；`AdminUsersStore.List`。
- **2026-03-23**：侧栏收敛，去掉「风控与合规」占位菜单；`/reconcile` 对账中心对接 `GET /v1/admin/reconcile/day`（trade `AdminDayOverview`）；`/settlement` 说明页；`order.proto` 增加 `AdminDayOverview`。
- **2026-03-23**：`/refunds` 退款与差错 MVP 说明页；`/ops` 运维监控对接 `GET /health`；网关 `GET /health`。
- **2026-03-23**：`/channel-health` 通道监控 MVP（路由汇总 + 通道列表只读）。
- **2026-03-23**：`/pay-orders` Pay订单列表与 `GET /v1/admin/pay_orders`；trade `AdminListPayOrders` RPC；菜单表补充代付产品路由说明。
- **2026-03-23**：`/routing` 路由策略页与 `GET /v1/admin/routing/summary`；`/stats` 与 `GET /v1/admin/stats/overview`。
- **2026-03-22**：`/merchant-products` 对接支付产品与通道绑定 API；页面 `PayProductsPage.vue`。
- 文档与前端菜单随 `adminMenu.ts` 同步维护；变更时请更新本节或提交说明。
