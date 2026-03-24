export default {
  meta: {
    title: 'Gloop Pay — 聚合支付基础设施',
    description: '稳定、高并发的支付中台与结算能力',
  },
  nav: {
    home: '首页',
    products: '产品与服务',
    docs: '开发者',
    about: '关于与合规',
  },
  brand: {
    name: 'Gloop Pay',
    tagline: '支付基础设施',
  },
  footer: {
    rights: '保留所有权利。',
    privacy: '隐私',
    terms: '条款',
    contact: '联系商务',
  },
  home: {
    badge: '聚合支付中台',
    heroTitle: '稳定 · 高并发 · 多渠道',
    heroLead:
      '统一接入多个支付通道，提供下单、查单、回调、结算与智能路由，为全球业务提供可扩展的支付基础设施。',
    ctaPrimary: '开始集成',
    ctaSecondary: '了解产品',
    features: {
      collect: {
        title: '快捷收单',
        desc: '扫码、H5、App 拉起，统一订单模型与幂等策略，提升支付成功率。',
      },
      payout: {
        title: '代付与结算',
        desc: '支持多种结算节奏，资金流水可审计，余额变更可追溯。',
      },
      routing: {
        title: '智能路由',
        desc: '按权重与金额段分流，熔断故障通道，保障业务连续性。',
      },
    },
    partners: '合作渠道与生态',
    trust: {
      title: '为增长而设计',
      desc: '从初创团队到规模化企业，一致的 API 体验与可观测性。',
    },
    stats: {
      uptime: '可用性目标',
      latency: '接口响应',
      regions: '多区域部署',
      uptimeVal: '99.95%',
      latencyVal: '< 120ms',
      regionsVal: '多区域',
    },
    regions: {
      title: '服务区域',
      lead: '覆盖中国、美国、日本、印度、巴西等主要经济体，对接本地支付工具、卡组与银行网络。',
      countries: {
        cn: {
          name: '中国',
          paymentsLabel: '支付与钱包',
          payments: '支付宝、微信支付、银联云闪付、京东支付、美团支付等',
          banksLabel: '银行（示例）',
          banks: '工商银行、建设银行、中国银行、招商银行、农业银行等',
        },
        us: {
          name: '美国',
          paymentsLabel: '支付与钱包',
          payments: 'PayPal、Stripe、Apple Pay、Google Pay、Venmo、Cash App 等',
          banksLabel: '银行（示例）',
          banks: '摩根大通、美国银行、富国银行、花旗银行、第一资本等',
        },
        jp: {
          name: '日本',
          paymentsLabel: '支付与钱包',
          payments: 'PayPay、LINE Pay、乐天支付、Mercari Pay、便利店（Konbini）支付等',
          banksLabel: '银行（示例）',
          banks: '三菱 UFJ、三井住友、瑞穗、理索纳、邮储银行等',
        },
        in: {
          name: '印度',
          paymentsLabel: '支付与钱包',
          payments: 'UPI（PhonePe、Google Pay、Paytm）、Razorpay、Paytm 钱包、Amazon Pay 等',
          banksLabel: '银行（示例）',
          banks: '印度国家银行、HDFC、ICICI、Axis、Kotak Mahindra 等',
        },
        br: {
          name: '巴西',
          paymentsLabel: '支付与钱包',
          payments: 'Pix、Boleto、PicPay、Mercado Pago、Nubank 等',
          banksLabel: '银行（示例）',
          banks: 'Itaú Unibanco、Bradesco、巴西银行、Santander 巴西、Caixa 等',
        },
      },
    },
    crypto: {
      title: '加密货币支付',
      lead:
        '在法币收单之外，支持数字资产支付：统一订单生命周期、确认数策略、结算与对账，面向全球化商户。',
      items: {
        onchain: {
          title: '链上与 Layer 2',
          desc: '支持主流公链与二层网络，可配置确认深度与链上监控，降低重组与拥堵风险。',
        },
        stablecoin: {
          title: '稳定币与报价',
          desc: 'USDT、USDC 等经审批的资产，支持实时报价、滑点与金库策略，便于财务核算。',
        },
        compliance: {
          title: '风控与合规',
          desc: '地址筛查、Travel Rule 对接准备与分司法辖区策略，帮助降低合规与反洗钱风险。',
        },
        experience: {
          title: '接入与运维',
          desc: '统一 API 创建加密收银会话，分级 Webhook 通知确认进度，并导出对账与报表。',
        },
      },
      disclaimer: '加密能力受地区、牌照与风控策略限制；具体资产与网络以审批为准。',
    },
  },
  products: {
    title: '产品与服务',
    lead: '面向不同业务场景的支付与结算解决方案。',
    scenarios: {
      ecommerce: {
        title: '电商',
        desc: '高并发订单、支付成功率优化、异步通知可靠送达。',
      },
      gaming: {
        title: '游戏与数字内容',
        desc: '高频小额、风控拦截、通道自动切换。',
      },
      retail: {
        title: '线下零售',
        desc: '扫码收款、统一对账、清结算透明可追溯。',
      },
    },
    settlement: {
      title: '代付 / 结算',
      lead: '支持多种结算节奏，资金流水与余额变更在同一事务内完成。',
      d0: { label: 'D+0', desc: '支付成功后快速入账。' },
      t1: { label: 'T+1', desc: '隔日结算，便于对账与风控。' },
      recon: { label: '对账', desc: '按日账单，差错订单可追踪。' },
    },
  },
  docs: {
    title: '开发者中心',
    lead: '快速集成、SDK 与文档入口。',
    quick: {
      title: '快速集成',
      step1: { title: '1. 创建订单', desc: '调用 CreateOrder，获取 order_no。' },
      step2: { title: '2. 跳转收银台', desc: '用户完成支付，平台异步处理回调。' },
      step3: { title: '3. Webhook', desc: '消息队列与消费者重试，可靠通知商户。' },
    },
    sdk: {
      title: 'SDK 与示例',
      go: 'Go（示例）',
      php: 'PHP（示例）',
      java: 'Java（示例）',
      python: 'Python（示例）',
    },
    online: {
      title: '在线文档',
      lead: '可接入 OpenAPI、Swagger / Redoc 或自建文档站。',
      swagger: 'Swagger',
      gitbook: 'GitBook',
    },
  },
  about: {
    title: '关于我们 & 合规',
    lead: '公司介绍、安全与合规、服务协议与隐私政策。',
    company: {
      title: '公司介绍',
      desc: '我们致力于为商户提供稳定、高性能的支付基础设施，帮助业务快速完成渠道接入与资金结算。',
    },
    security: {
      title: '安全与合规',
      desc: '敏感信息加密存储、操作审计与风险控制，可对接 PCI-DSS 等行业安全要求。',
      pci: 'PCI-DSS',
      privacy: '隐私政策',
      terms: '服务协议',
    },
  },
}
