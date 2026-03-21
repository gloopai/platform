import { createRouter, createWebHistory } from 'vue-router'

import AboutPage from './views/AboutPage.vue'
import DocsPage from './views/DocsPage.vue'
import HomePage from './views/HomePage.vue'
import ProductsPage from './views/ProductsPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomePage },
    { path: '/products', component: ProductsPage },
    { path: '/docs', component: DocsPage },
    { path: '/about', component: AboutPage },
  ],
})
