import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './views/LoginPage.vue'
import MerchantLayout from './views/MerchantLayout.vue'
import ConsolePage from './views/pages/ConsolePage.vue'
import DevelopersPage from './views/pages/DevelopersPage.vue'
import FinancePage from './views/pages/FinancePage.vue'
import ModulePlaceholderPage from './views/pages/ModulePlaceholderPage.vue'
import ProductsPage from './views/pages/ProductsPage.vue'
import TransactionsPayinPage from './views/pages/TransactionsPayinPage.vue'
import TransactionsPayoutPage from './views/pages/TransactionsPayoutPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginPage },
    {
      path: '/',
      component: MerchantLayout,
      children: [
        { path: '', redirect: '/console' },
        { path: 'console', component: ConsolePage },
        { path: 'transactions-payin', component: TransactionsPayinPage },
        { path: 'transactions-payout', component: TransactionsPayoutPage },
        { path: 'finance', component: FinancePage },
        { path: 'products', component: ProductsPage },
        { path: 'account', component: ModulePlaceholderPage },
        { path: 'developers', component: DevelopersPage },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.path === '/login') return true
  const tok = localStorage.getItem('merchant_token')
  if (!tok) return '/login'
  return true
})
