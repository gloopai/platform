<template>
  <RouterView />
  <UiToastHost />
  <UiDialogHost />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import UiDialogHost from '../../../shared/ui/dialog/UiDialogHost.vue'
import UiToastHost from '../../../shared/ui/toast/UiToastHost.vue'
</script>
