import { createApp } from 'vue'

import App from './App.vue'
import { getInitialLocale, htmlLangFromLocale, i18n } from './i18n'
import { router } from './router'
import './style.css'

document.documentElement.lang = htmlLangFromLocale(getInitialLocale())
document.documentElement.setAttribute('data-theme', 'gloopmono')

createApp(App).use(router).use(i18n).mount('#app')
