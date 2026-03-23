# pay-platform

四方聚合支付相关单体/微服务与前端（管理台、商户台、收银台等）。

## 本地把主路径跑通

1. 初始化数据库：`bash backend/deploy/init_demo.sh`（按脚本内说明配置 MySQL 账号）  
2. 启动：`bash dev-up.sh`（网关默认 `http://127.0.0.1:8080`，收银台 `http://127.0.0.1:5174`）  
3. 按文档从 **下单 → 收银台 → 模拟回调 → 查单** 走一遍：[`docs/端到端联调一遍.md`](docs/端到端联调一遍.md)

更多：[`docs/开发计划.md`](docs/开发计划.md)、[`docs/MVP与后续迭代.md`](docs/MVP与后续迭代.md)。
