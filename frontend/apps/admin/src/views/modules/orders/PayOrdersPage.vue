<template>
  <OrdersPageBase
    title="代收订单"
    description="跨商户检索代收订单（payin_orders）"
    endpoint="/v1/admin/payin_orders"
  />
</template>

<script setup lang="ts">
import OrdersPageBase from './OrdersPage.vue'
</script>
