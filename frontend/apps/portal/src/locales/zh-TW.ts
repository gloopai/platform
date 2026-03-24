export default {
  meta: {
    title: 'Gloop Pay — 聚合支付基礎設施',
    description: '穩定、高並發的支付中台與結算能力',
  },
  nav: {
    home: '首頁',
    products: '產品與服務',
    docs: '開發者',
    about: '關於與合規',
  },
  brand: {
    name: 'Gloop Pay',
    tagline: '支付基礎設施',
  },
  footer: {
    rights: '保留所有權利。',
    privacy: '隱私',
    terms: '條款',
    contact: '聯繫商務',
  },
  home: {
    badge: '聚合支付中台',
    heroTitle: '穩定 · 高並發 · 多渠道',
    heroLead:
      '統一接入多個支付通道，提供下單、查單、回調、結算與智能路由，為全球業務提供可擴展的支付基礎設施。',
    ctaPrimary: '開始整合',
    ctaSecondary: '了解產品',
    features: {
      collect: {
        title: '快捷收單',
        desc: '掃碼、H5、App 拉起，統一訂單模型與冪等策略，提升支付成功率。',
      },
      payout: {
        title: '代付與結算',
        desc: '支援多種結算節奏，資金流水可審計，餘額變更可追溯。',
      },
      routing: {
        title: '智能路由',
        desc: '按權重與金額段分流，熔斷故障通道，保障業務連續性。',
      },
    },
    partners: '合作渠道與生態',
    trust: {
      title: '為增長而設計',
      desc: '從初創團隊到規模化企業，一致的 API 體驗與可觀測性。',
    },
    stats: {
      uptime: '可用性目標',
      latency: '介面回應',
      regions: '多區域部署',
      uptimeVal: '99.95%',
      latencyVal: '< 120ms',
      regionsVal: '多區域',
    },
    regions: {
      title: '服務區域',
      lead: '涵蓋中國、美國、日本、印度、巴西等主要經濟體，對接在地支付工具、卡組織與銀行網路。',
      countries: {
        cn: {
          name: '中國',
          paymentsLabel: '支付與錢包',
          payments: '支付寶、微信支付、銀聯雲閃付、京東支付、美團支付等',
          banksLabel: '銀行（示例）',
          banks: '工商銀行、建設銀行、中國銀行、招商銀行、農業銀行等',
        },
        us: {
          name: '美國',
          paymentsLabel: '支付與錢包',
          payments: 'PayPal、Stripe、Apple Pay、Google Pay、Venmo、Cash App 等',
          banksLabel: '銀行（示例）',
          banks: '摩根大通、美國銀行、富國銀行、花旗銀行、第一資本等',
        },
        jp: {
          name: '日本',
          paymentsLabel: '支付與錢包',
          payments: 'PayPay、LINE Pay、樂天支付、Mercari Pay、便利商店（Konbini）支付等',
          banksLabel: '銀行（示例）',
          banks: '三菱 UFJ、三井住友、瑞穗、理索納、郵儲銀行等',
        },
        in: {
          name: '印度',
          paymentsLabel: '支付與錢包',
          payments: 'UPI（PhonePe、Google Pay、Paytm）、Razorpay、Paytm 錢包、Amazon Pay 等',
          banksLabel: '銀行（示例）',
          banks: '印度國家銀行、HDFC、ICICI、Axis、Kotak Mahindra 等',
        },
        br: {
          name: '巴西',
          paymentsLabel: '支付與錢包',
          payments: 'Pix、Boleto、PicPay、Mercado Pago、Nubank 等',
          banksLabel: '銀行（示例）',
          banks: 'Itaú Unibanco、Bradesco、巴西銀行、Santander 巴西、Caixa 等',
        },
      },
    },
  },
  products: {
    title: '產品與服務',
    lead: '面向不同業務場景的支付與結算解決方案。',
    scenarios: {
      ecommerce: {
        title: '電商',
        desc: '高並發訂單、支付成功率優化、非同步通知可靠送達。',
      },
      gaming: {
        title: '遊戲與數位內容',
        desc: '高頻小額、風控攔截、通道自動切換。',
      },
      retail: {
        title: '線下零售',
        desc: '掃碼收款、統一對帳、清結算透明可追溯。',
      },
    },
    settlement: {
      title: '代付 / 結算',
      lead: '支援多種結算節奏，資金流水與餘額變更在同一事務內完成。',
      d0: { label: 'D+0', desc: '支付成功後快速入帳。' },
      t1: { label: 'T+1', desc: '隔日結算，便於對帳與風控。' },
      recon: { label: '對帳', desc: '按日帳單，差錯訂單可追蹤。' },
    },
  },
  docs: {
    title: '開發者中心',
    lead: '快速整合、SDK 與文件入口。',
    quick: {
      title: '快速整合',
      step1: { title: '1. 建立訂單', desc: '呼叫 CreateOrder，取得 order_no。' },
      step2: { title: '2. 跳轉收銀台', desc: '使用者完成支付，平台非同步處理回調。' },
      step3: { title: '3. Webhook', desc: '訊息佇列與消費者重試，可靠通知商戶。' },
    },
    sdk: {
      title: 'SDK 與範例',
      go: 'Go（範例）',
      php: 'PHP（範例）',
      java: 'Java（範例）',
      python: 'Python（範例）',
    },
    online: {
      title: '線上文件',
      lead: '可接入 OpenAPI、Swagger / Redoc 或自建文件站。',
      swagger: 'Swagger',
      gitbook: 'GitBook',
    },
  },
  about: {
    title: '關於我們與合規',
    lead: '公司介紹、安全與合規、服務協議與隱私政策。',
    company: {
      title: '公司介紹',
      desc: '我們致力於為商戶提供穩定、高效能的支付基礎設施，幫助業務快速完成渠道接入與資金結算。',
    },
    security: {
      title: '安全與合規',
      desc: '敏感資訊加密儲存、操作審計與風險控制，可對接 PCI-DSS 等行業安全要求。',
      pci: 'PCI-DSS',
      privacy: '隱私政策',
      terms: '服務協議',
    },
  },
}
