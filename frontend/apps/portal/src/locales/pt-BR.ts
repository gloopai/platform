export default {
  meta: {
    title: 'Gloop Pay — Infraestrutura de pagamentos',
    description: 'Adquirência e liquidação confiáveis e de alto desempenho',
  },
  nav: {
    home: 'Início',
    products: 'Produtos',
    docs: 'Desenvolvedores',
    about: 'Sobre e conformidade',
  },
  brand: {
    name: 'Gloop Pay',
    tagline: 'Infraestrutura de pagamentos',
  },
  footer: {
    rights: 'Todos os direitos reservados.',
    privacy: 'Privacidade',
    terms: 'Termos',
    contact: 'Fale com vendas',
  },
  home: {
    badge: 'Plataforma unificada de pagamentos',
    heroTitle: 'Estável · Alta concorrência · Multicanal',
    heroLead:
      'Conecte vários meios de pagamento em uma única integração—pedidos, consultas, callbacks, liquidação e roteamento inteligente em escala global.',
    ctaPrimary: 'Começar a integrar',
    ctaSecondary: 'Ver produtos',
    features: {
      collect: {
        title: 'Captura rápida',
        desc: 'QR, H5 e fluxos no app com modelo de pedido unificado e APIs idempotentes.',
      },
      payout: {
        title: 'Pagamentos e liquidação',
        desc: 'Ciclos de liquidação flexíveis, razão auditável e saldos rastreáveis.',
      },
      routing: {
        title: 'Roteamento inteligente',
        desc: 'Roteamento por peso e faixa de valor, com disjuntores para canais indisponíveis.',
      },
    },
    partners: 'Canais e ecossistema',
    trust: {
      title: 'Feito para crescer',
      desc: 'De startups a grandes empresas—APIs consistentes e observabilidade de ponta a ponta.',
    },
    stats: {
      uptime: 'Meta de disponibilidade',
      latency: 'Latência da API',
      regions: 'Pronto para várias regiões',
      uptimeVal: '99,95%',
      latencyVal: '< 120ms',
      regionsVal: 'Multirregião',
    },
    regions: {
      title: 'Regiões atendidas',
      lead:
        'Conectamos comerciantes às redes locais de pagamento na China, Estados Unidos, Japão, Índia e Brasil—incluindo carteiras, bandeiras e bancos.',
      countries: {
        cn: {
          name: 'China',
          paymentsLabel: 'Pagamentos e carteiras',
          payments: 'Alipay, WeChat Pay, UnionPay QuickPass, JD Pay, Meituan Pay',
          banksLabel: 'Bancos (exemplos)',
          banks: 'ICBC, China Construction Bank, Bank of China, China Merchants Bank, Agricultural Bank of China',
        },
        us: {
          name: 'Estados Unidos',
          paymentsLabel: 'Pagamentos e carteiras',
          payments: 'PayPal, Stripe, Apple Pay, Google Pay, Venmo, Cash App',
          banksLabel: 'Bancos (exemplos)',
          banks: 'JPMorgan Chase, Bank of America, Wells Fargo, Citibank, Capital One',
        },
        jp: {
          name: 'Japão',
          paymentsLabel: 'Pagamentos e carteiras',
          payments: 'PayPay, LINE Pay, Rakuten Pay, Mercari Pay, pagamentos em kombini (lojas de conveniência)',
          banksLabel: 'Bancos (exemplos)',
          banks: 'MUFG, SMBC, Mizuho, Resona Bank, Japan Post Bank',
        },
        in: {
          name: 'Índia',
          paymentsLabel: 'Pagamentos e carteiras',
          payments: 'UPI (PhonePe, Google Pay, Paytm), Razorpay, carteira Paytm, Amazon Pay',
          banksLabel: 'Bancos (exemplos)',
          banks: 'SBI, HDFC Bank, ICICI Bank, Axis Bank, Kotak Mahindra Bank',
        },
        br: {
          name: 'Brasil',
          paymentsLabel: 'Pagamentos e carteiras',
          payments: 'Pix, boleto bancário, PicPay, Mercado Pago, Nubank',
          banksLabel: 'Bancos (exemplos)',
          banks: 'Itaú Unibanco, Bradesco, Banco do Brasil, Santander Brasil, Caixa Econômica Federal',
        },
      },
    },
    crypto: {
      title: 'Pagamentos em criptomoedas',
      lead:
        'Aceite ativos digitais junto com fiat—ciclo de pedido unificado, confirmações, liquidação e relatórios para comerciantes globais.',
      items: {
        onchain: {
          title: 'On-chain e Layer 2',
          desc: 'Redes principais e L2 com profundidade de confirmação configurável e monitoramento.',
        },
        stablecoin: {
          title: 'Stablecoins e preços',
          desc: 'USDT, USDC e outros ativos aprovados com cotações em tempo real, slippage e políticas de tesouraria.',
        },
        compliance: {
          title: 'Risco e conformidade',
          desc: 'Triagem de endereços, preparação para travel rule e políticas por jurisdição.',
        },
        experience: {
          title: 'API e operações',
          desc: 'Uma API para sessões de checkout cripto, webhooks por nível de confirmação e exportações de conciliação.',
        },
      },
      disclaimer:
        'Recursos de cripto dependem de região, licença e política de risco; ativos e redes sujeitos à aprovação.',
    },
  },
  products: {
    title: 'Produtos e serviços',
    lead: 'Soluções de pagamento e liquidação para diferentes cenários de negócio.',
    scenarios: {
      ecommerce: {
        title: 'E-commerce',
        desc: 'Alto volume de pedidos, otimização de conversão e notificações assíncronas confiáveis.',
      },
      gaming: {
        title: 'Jogos e digital',
        desc: 'Micropagamentos de alta frequência, controles de risco e failover automático de canal.',
      },
      retail: {
        title: 'Varejo físico',
        desc: 'QR na loja, conciliação unificada e liquidação transparente.',
      },
    },
    settlement: {
      title: 'Pagamentos / liquidação',
      lead: 'Cadência de liquidação flexível; razão e saldo atualizados em um único fluxo transacional.',
      d0: { label: 'D+0', desc: 'Crédito rápido após pagamento bem-sucedido.' },
      t1: { label: 'T+1', desc: 'Liquidação no dia seguinte para conciliação e risco.' },
      recon: { label: 'Conciliação', desc: 'Extratos diários com exceções rastreáveis.' },
    },
  },
  docs: {
    title: 'Central do desenvolvedor',
    lead: 'Integre rapidamente—SDKs, exemplos e documentação.',
    quick: {
      title: 'Integração rápida',
      step1: { title: '1. Criar pedido', desc: 'Chame CreateOrder para obter order_no.' },
      step2: { title: '2. Checkout hospedado', desc: 'O cliente paga; a plataforma processa callbacks de forma assíncrona.' },
      step3: { title: '3. Webhooks', desc: 'Entrega com fila e novas tentativas até o seu endpoint.' },
    },
    sdk: {
      title: 'SDKs e exemplos',
      go: 'Go (exemplo)',
      php: 'PHP (exemplo)',
      java: 'Java (exemplo)',
      python: 'Python (exemplo)',
    },
    online: {
      title: 'Documentação',
      lead: 'OpenAPI, Swagger / Redoc ou seu próprio site de documentação.',
      swagger: 'Swagger',
      gitbook: 'GitBook',
    },
  },
  about: {
    title: 'Sobre e conformidade',
    lead: 'Perfil da empresa, segurança, termos e privacidade.',
    company: {
      title: 'Empresa',
      desc: 'Construímos infraestrutura de pagamentos confiável e de alto desempenho para integrar canais e liquidar fundos com mais rapidez.',
    },
    security: {
      title: 'Segurança e conformidade',
      desc: 'Criptografia em repouso, trilhas de auditoria e controles de risco alinhados a padrões como PCI-DSS.',
      pci: 'PCI-DSS',
      privacy: 'Política de privacidade',
      terms: 'Termos de serviço',
    },
  },
}
