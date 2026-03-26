import { createApp } from 'vue'
import { router } from './router'
import './style.css'
import App from './App.vue'

document.documentElement.setAttribute('data-theme', 'gloopmono')

createApp(App).use(router).mount('#app')
