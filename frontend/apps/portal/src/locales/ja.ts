export default {
  meta: {
    title: 'Gloop Pay — 決済インフラ',
    description: '高信頼・高スループットの収単と決済',
  },
  nav: {
    home: 'ホーム',
    products: '製品・サービス',
    docs: '開発者',
    about: '会社・コンプライアンス',
  },
  brand: {
    name: 'Gloop Pay',
    tagline: '決済インフラ',
  },
  footer: {
    rights: '無断転載を禁じます。',
    privacy: 'プライバシー',
    terms: '規約',
    contact: '営業へのお問い合わせ',
  },
  home: {
    badge: '統合決済プラットフォーム',
    heroTitle: '安定 · 高並列 · マルチチャネル',
    heroLead:
      '複数の決済チャネルを一つの連携で。注文・照会・コールバック・決済・インテリジェントルーティングをグローバル規模で提供します。',
    ctaPrimary: '連携を始める',
    ctaSecondary: '製品を見る',
    features: {
      collect: {
        title: '迅速な収単',
        desc: 'QR、H5、アプリ内決済。統一された注文モデルと冪等 API。',
      },
      payout: {
        title: '送金・精算',
        desc: '柔軟な精算サイクル、監査可能な台帳、追跡可能な残高変更。',
      },
      routing: {
        title: 'スマートルーティング',
        desc: '重み付けと金額帯による振り分け、不健全チャネルのサーキットブレーカー。',
      },
    },
    partners: 'チャネルとエコシステム',
    trust: {
      title: '成長のために',
      desc: 'スタートアップから大企業まで、一貫した API とオブザーバビリティ。',
    },
    stats: {
      uptime: '可用性目標',
      latency: 'API レイテンシ',
      regions: 'マルチリージョン',
      uptimeVal: '99.95%',
      latencyVal: '< 120ms',
      regionsVal: 'マルチリージョン',
    },
    regions: {
      title: 'サービス提供地域',
      lead:
        '中国・米国・日本・インド・ブラジルの主要経済圏で、現地の決済手段と銀行ネットワークに接続します。',
      countries: {
        cn: {
          name: '中国',
          paymentsLabel: '決済・ウォレット',
          payments: 'Alipay、WeChat Pay、銀聯（UnionPay）クイックパス、JD Pay、美団（Meituan）Pay など',
          banksLabel: '銀行（例）',
          banks: '中国工商銀行、中国建設銀行、中国銀行、招商銀行、中国農業銀行 など',
        },
        us: {
          name: 'アメリカ合衆国',
          paymentsLabel: '決済・ウォレット',
          payments: 'PayPal、Stripe、Apple Pay、Google Pay、Venmo、Cash App',
          banksLabel: '銀行（例）',
          banks: 'JPMorgan Chase、Bank of America、Wells Fargo、Citibank、Capital One',
        },
        jp: {
          name: '日本',
          paymentsLabel: '決済・ウォレット',
          payments: 'PayPay、LINE Pay、楽天ペイ、メルペイ、コンビニ（Konbini）決済 など',
          banksLabel: '銀行（例）',
          banks: '三菱UFJ、三井住友、みずほ、りそな、ゆうちょ銀行',
        },
        in: {
          name: 'インド',
          paymentsLabel: '決済・ウォレット',
          payments: 'UPI（PhonePe、Google Pay、Paytm）、Razorpay、Paytm ウォレット、Amazon Pay など',
          banksLabel: '銀行（例）',
          banks: 'SBI、HDFC Bank、ICICI Bank、Axis Bank、Kotak Mahindra Bank',
        },
        br: {
          name: 'ブラジル',
          paymentsLabel: '決済・ウォレット',
          payments: 'Pix、Boleto bancário、PicPay、Mercado Pago、Nubank',
          banksLabel: '銀行（例）',
          banks: 'Itaú Unibanco、Bradesco、Banco do Brasil、Santander Brasil、Caixa Econômica Federal',
        },
      },
    },
    crypto: {
      title: '暗号資産決済',
      lead:
        '法定通貨に加え、デジタル資産の受け入れを統一された注文ライフサイクルで提供—確認、精算、レポートまでグローバル向けに設計。',
      items: {
        onchain: {
          title: 'オンチェーンと L2',
          desc: '主要チェーンと L2 をサポート。確認数の設定やモニタリングでリスクを抑制。',
        },
        stablecoin: {
          title: 'ステーブルコインと価格',
          desc: 'USDT、USDC など承認済み資産にリアルタイムレート、スリッページ管理、トレジャリー方針。',
        },
        compliance: {
          title: 'リスクとコンプライアンス',
          desc: 'アドレススクリーニング、Travel Rule への準備、法域に応じたポリシーで露出を低減。',
        },
        experience: {
          title: '開発と運用',
          desc: '暗号チェックアウトを単一 API で作成、段階的 Webhook、照合用エクスポート。',
        },
      },
      disclaimer: '暗号資産機能は地域・ライセンス・リスク方針により異なります。対応資産とネットワークは承認制です。',
    },
  },
  products: {
    title: '製品・サービス',
    lead: 'さまざまな業務シーン向けの決済・精算ソリューション。',
    scenarios: {
      ecommerce: {
        title: 'Eコマース',
        desc: '高い注文処理能力、コンバージョン最適化、信頼できる非同期通知。',
      },
      gaming: {
        title: 'ゲーム・デジタル',
        desc: '高頻度の少額決済、リスク管理、自動チャネル切替。',
      },
      retail: {
        title: '実店舗',
        desc: '店頭 QR、統合された照合、透明な精算。',
      },
    },
    settlement: {
      title: '送金 / 精算',
      lead: '柔軟な精算サイクル。台帳と残高を一つのトランザクションで更新。',
      d0: { label: 'D+0', desc: '決済成功後の迅速な入金。' },
      t1: { label: 'T+1', desc: '翌営業日精算。照合とリスク管理に適す。' },
      recon: { label: '照合', desc: '日次明細、例外注文の追跡。' },
    },
  },
  docs: {
    title: '開発者ハブ',
    lead: 'すばやく連携—SDK、サンプル、ドキュメント。',
    quick: {
      title: 'クイック連携',
      step1: { title: '1. 注文作成', desc: 'CreateOrder を呼び出し order_no を取得。' },
      step2: { title: '2. ホステッドチェックアウト', desc: '顧客が支払い、プラットフォームが非同期でコールバック処理。' },
      step3: { title: '3. Webhook', desc: 'キューと再試行でエンドポイントへ確実に配信。' },
    },
    sdk: {
      title: 'SDK とサンプル',
      go: 'Go（サンプル）',
      php: 'PHP（サンプル）',
      java: 'Java（サンプル）',
      python: 'Python（サンプル）',
    },
    online: {
      title: 'ドキュメント',
      lead: 'OpenAPI、Swagger / Redoc、または独自のドキュメントサイト。',
      swagger: 'Swagger',
      gitbook: 'GitBook',
    },
  },
  about: {
    title: '会社・コンプライアンス',
    lead: '会社概要、セキュリティ、利用規約とプライバシー。',
    company: {
      title: '会社概要',
      desc: '安定した高性能の決済インフラを提供し、チャネル接続と資金精算を迅速化します。',
    },
    security: {
      title: 'セキュリティ・コンプライアンス',
      desc: '保存時の暗号化、監査ログ、リスク管理。PCI-DSS 等の業界基準に整合。',
      pci: 'PCI-DSS',
      privacy: 'プライバシーポリシー',
      terms: '利用規約',
    },
  },
}
