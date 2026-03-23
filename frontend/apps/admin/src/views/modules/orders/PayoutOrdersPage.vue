<template>
  <OrdersPageBase
    title="代付订单"
    description="跨商户检索代付订单（payout_orders）"
    endpoint="/v1/admin/payout_orders"
  />
</template>

<script setup lang="ts">
import OrdersPageBase from './OrdersPage.vue'
</script>
