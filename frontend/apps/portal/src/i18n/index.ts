import { createI18n } from 'vue-i18n'

import en from '../locales/en'
import hiIN from '../locales/hi-IN'
import ja from '../locales/ja'
import ptBR from '../locales/pt-BR'
import zhCN from '../locales/zh-CN'
import zhTW from '../locales/zh-TW'

const STORAGE_KEY = 'portal-locale'

const SUPPORTED = ['en', 'zh-CN', 'zh-TW', 'ja', 'pt-BR', 'hi-IN'] as const

export type PortalLocale = (typeof SUPPORTED)[number]

/** 首次访问默认英文；仅当 localStorage 中有用户曾选择的语言时才沿用。 */
export function getInitialLocale(): PortalLocale {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved && (SUPPORTED as readonly string[]).includes(saved)) {
      return saved as PortalLocale
    }
  } catch {
    /* ignore */
  }
  return 'en'
}

export function persistLocale(locale: string) {
  try {
    localStorage.setItem(STORAGE_KEY, locale)
  } catch {
    /* ignore */
  }
}

export function htmlLangFromLocale(locale: string): string {
  switch (locale) {
    case 'zh-TW':
      return 'zh-Hant'
    case 'zh-CN':
      return 'zh-Hans'
    case 'ja':
      return 'ja'
    case 'pt-BR':
      return 'pt-BR'
    case 'hi-IN':
      return 'hi-IN'
    default:
      return 'en'
  }
}

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    'zh-CN': zhCN,
    'zh-TW': zhTW,
    ja,
    'pt-BR': ptBR,
    'hi-IN': hiIN,
  },
})
