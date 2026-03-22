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
| `/stats` | 系统概览 | 平台交易额、笔数、成功率、通道可用率、待办与告警；**部分指标待接入** |

### 2.2 商户与接入

| 路径 | 页面 | 说明 |
|------|------|------|
| `/merchants` | 商户管理 | **已对接**：网关 `/v1/admin/merchants` 列表/创建/更新 |
| `/merchant-products` | 支付产品与上游通道 | **已对接**：`pay_products` CRUD 与 `pay_product_channels` 绑定（权重、启用）；见 §4 |

### 2.3 通道与路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/channels` | 通道管理 | **已对接**：`/v1/admin/channels` 列表/创建/更新 |
| `/routing` | 路由策略 | 按金额/地区/商户等级/权重分配通道；**占位** |
| `/channel-health` | 通道监控 | 成功率、延迟、熔断、告警；**占位** |

### 2.4 交易与资金

| 路径 | 页面 | 说明 |
|------|------|------|
| `/orders` | 全站订单 | 跨商户订单检索、导出、异常处理；**占位** |
| `/refunds` | 退款与差错 | 退款审核、差错单、长短款；**占位** |
| `/reconcile` | 对账中心 | 对账批次、差异处理；**占位** |
| `/settlement` | 结算与提现 | 结算周期、提现审核、打款；**占位** |

### 2.5 风控与合规

| 路径 | 页面 | 说明 |
|------|------|------|
| `/risk` | 风控规则 | 限额、黑名单、评分策略；**占位** |
| `/audit` | 运营与审计 | 运营动作、审计日志、RBAC 规划；**前端说明页** |
| `/notifications` | 公告与通知 | 对商户公告、维护窗口；**占位** |

### 2.6 系统与运维

| 路径 | 页面 | 说明 |
|------|------|------|
| `/system` | 系统管理 | 管理员、角色、系统参数；**占位** |
| `/ops` | 运维监控 | 服务健康、QPS、链路；**占位** |

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

其余菜单对应能力 **尚未实现**；前端以 `ModulePlaceholderPage` + `adminMenu.ts` 中 `adminPlaceholderMeta` 说明规划与待接入 API 提示。

---

## 5. 前端实现说明

开发约定（鉴权请求封装、按模块分目录与组件拆分）见 [**管理端前端开发规范**](./管理端前端开发规范.md)。

| 文件 | 作用 |
|------|------|
| `src/adminMenu.ts` | 侧栏菜单结构、面包屑标题、占位页文案 |
| `src/views/AdminLayout.vue` | 多级侧栏、折叠态悬浮子菜单 |
| `src/views/modules/pay-products/` | 支付产品与上游通道绑定（`PayProductsPage.vue` + 子组件） |
| `src/views/modules/channels/` | 通道管理（`ChannelsPage.vue` + `ChannelList` / `ChannelFormCard` 等） |
| `src/views/pages/ModulePlaceholderPage.vue` | 通用占位页（读 `adminPlaceholderMeta`） |
| `src/router.ts` | 路由注册 |

---

## 6. 迭代建议（Roadmap）

1. **数据与报表**：总览大盘对接 `/v1/admin/dashboard` 或直连统计服务。
2. **订单全站**：`orders` 与 trade 服务查询接口，支持导出与权限。
3. **路由与通道**：路由规则存储与执行引擎；通道健康时序指标。
4. **对账与结算**：对账任务表、文件存储、结算单状态机。
5. **RBAC 与审计**：管理员表、角色、菜单与操作日志表。

---

## 7. 修订记录

- **2026-03-22**：`/merchant-products` 对接支付产品与通道绑定 API；页面 `PayProductsPage.vue`。
- 文档与前端菜单随 `adminMenu.ts` 同步维护；变更时请更新本节或提交说明。
