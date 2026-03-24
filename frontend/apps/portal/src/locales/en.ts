export default {
  meta: {
    title: 'Gloop Pay — Payment infrastructure',
    description: 'Reliable, high-throughput acquiring and settlement',
  },
  nav: {
    home: 'Home',
    products: 'Products',
    docs: 'Developers',
    about: 'About & compliance',
  },
  brand: {
    name: 'Gloop Pay',
    tagline: 'Payment infrastructure',
  },
  footer: {
    rights: 'All rights reserved.',
    privacy: 'Privacy',
    terms: 'Terms',
    contact: 'Contact sales',
  },
  home: {
    badge: 'Unified payment platform',
    heroTitle: 'Stable · High concurrency · Multi-channel',
    heroLead:
      'Connect multiple payment channels through one integration—orders, queries, callbacks, settlement, and intelligent routing built for global scale.',
    ctaPrimary: 'Start integrating',
    ctaSecondary: 'Explore products',
    features: {
      collect: {
        title: 'Fast acquiring',
        desc: 'QR, H5, and in-app flows with a unified order model and idempotent APIs.',
      },
      payout: {
        title: 'Payouts & settlement',
        desc: 'Flexible settlement cycles with auditable ledgers and traceable balance changes.',
      },
      routing: {
        title: 'Smart routing',
        desc: 'Weight- and amount-based routing with circuit breakers for unhealthy channels.',
      },
    },
    partners: 'Channels & ecosystem',
    trust: {
      title: 'Built for growth',
      desc: 'From startups to enterprises—consistent APIs and observability across the stack.',
    },
    stats: {
      uptime: 'Availability target',
      latency: 'API latency',
      regions: 'Multi-region ready',
      uptimeVal: '99.95%',
      latencyVal: '< 120ms',
      regionsVal: 'Multi-region',
    },
    regions: {
      title: 'Global service regions',
      lead:
        'We connect merchants to local payment rails in China, the United States, Japan, India, and Brazil—covering major wallets, card schemes, and bank networks.',
      countries: {
        cn: {
          name: 'China',
          paymentsLabel: 'Payments & wallets',
          payments: 'Alipay, WeChat Pay, UnionPay QuickPass, JD Pay, Meituan Pay',
          banksLabel: 'Banks (examples)',
          banks: 'ICBC, China Construction Bank, Bank of China, China Merchants Bank, Agricultural Bank of China',
        },
        us: {
          name: 'United States',
          paymentsLabel: 'Payments & wallets',
          payments: 'PayPal, Stripe, Apple Pay, Google Pay, Venmo, Cash App',
          banksLabel: 'Banks (examples)',
          banks: 'JPMorgan Chase, Bank of America, Wells Fargo, Citibank, Capital One',
        },
        jp: {
          name: 'Japan',
          paymentsLabel: 'Payments & wallets',
          payments: 'PayPay, LINE Pay, Rakuten Pay, Mercari Pay, convenience store (Konbini) payments',
          banksLabel: 'Banks (examples)',
          banks: 'MUFG, SMBC, Mizuho, Resona Bank, Japan Post Bank',
        },
        in: {
          name: 'India',
          paymentsLabel: 'Payments & wallets',
          payments: 'UPI (PhonePe, Google Pay, Paytm), Razorpay, Paytm Wallet, Amazon Pay',
          banksLabel: 'Banks (examples)',
          banks: 'SBI, HDFC Bank, ICICI Bank, Axis Bank, Kotak Mahindra Bank',
        },
        br: {
          name: 'Brazil',
          paymentsLabel: 'Payments & wallets',
          payments: 'Pix, Boleto bancário, PicPay, Mercado Pago, Nubank',
          banksLabel: 'Banks (examples)',
          banks: 'Itaú Unibanco, Bradesco, Banco do Brasil, Santander Brasil, Caixa Econômica Federal',
        },
      },
    },
    crypto: {
      title: 'Cryptocurrency payments',
      lead:
        'Accept digital assets alongside fiat—a unified order lifecycle with confirmations, settlement options, and reporting built for global merchants.',
      items: {
        onchain: {
          title: 'On-chain & Layer 2',
          desc: 'Support for major networks and L2s with configurable confirmation depth and monitoring.',
        },
        stablecoin: {
          title: 'Stablecoins & pricing',
          desc: 'USDT, USDC, and other approved assets with real-time quotes, slippage controls, and treasury policies.',
        },
        compliance: {
          title: 'Risk & compliance',
          desc: 'Address screening, travel-rule readiness, and jurisdiction-aware controls to reduce exposure.',
        },
        experience: {
          title: 'Developer & ops',
          desc: 'One API for crypto checkout sessions, tiered webhooks for confirmations, and reconciliation exports.',
        },
      },
      disclaimer:
        'Cryptocurrency features depend on region, licence, and risk policy; supported assets and networks are subject to approval.',
    },
  },
  products: {
    title: 'Products & services',
    lead: 'Payment and settlement solutions for different business scenarios.',
    scenarios: {
      ecommerce: {
        title: 'E‑commerce',
        desc: 'High order volume, conversion optimization, and reliable async notifications.',
      },
      gaming: {
        title: 'Gaming & digital',
        desc: 'High-frequency micropayments, risk controls, and automatic channel failover.',
      },
      retail: {
        title: 'Offline retail',
        desc: 'In-store QR payments, unified reconciliation, transparent settlement.',
      },
    },
    settlement: {
      title: 'Payouts / settlement',
      lead: 'Flexible settlement cadence; ledger and balance updates in a single transactional flow.',
      d0: { label: 'D+0', desc: 'Fast crediting after successful payment.' },
      t1: { label: 'T+1', desc: 'Next-day settlement for reconciliation and risk.' },
      recon: { label: 'Reconciliation', desc: 'Daily statements with traceable exceptions.' },
    },
  },
  docs: {
    title: 'Developer hub',
    lead: 'Integrate quickly—SDKs, examples, and documentation.',
    quick: {
      title: 'Quick integration',
      step1: { title: '1. Create order', desc: 'Call CreateOrder to obtain order_no.' },
      step2: { title: '2. Hosted checkout', desc: 'Customer pays; platform processes callbacks asynchronously.' },
      step3: { title: '3. Webhooks', desc: 'Queue-backed delivery with retries to your endpoint.' },
    },
    sdk: {
      title: 'SDKs & samples',
      go: 'Go (sample)',
      php: 'PHP (sample)',
      java: 'Java (sample)',
      python: 'Python (sample)',
    },
    online: {
      title: 'Documentation',
      lead: 'OpenAPI, Swagger / Redoc, or your own docs site.',
      swagger: 'Swagger',
      gitbook: 'GitBook',
    },
  },
  about: {
    title: 'About & compliance',
    lead: 'Company profile, security, terms, and privacy.',
    company: {
      title: 'Company',
      desc: 'We build reliable, high-performance payment infrastructure so merchants can onboard channels and settle funds faster.',
    },
    security: {
      title: 'Security & compliance',
      desc: 'Encryption at rest, audit trails, and risk controls aligned with industry standards such as PCI-DSS.',
      pci: 'PCI-DSS',
      privacy: 'Privacy policy',
      terms: 'Terms of service',
    },
  },
}
