<template>
  <UiDrawer
    :model-value="modelValue"
    :title="title"
    :subtitle="subtitle"
    :max-width-class="maxWidthClass"
  >
    <slot />
    <template #footer>
      <slot name="footer" />
    </template>
  </UiDrawer>
</template>

<script setup lang="ts">
import UiDrawer from '../../../../shared/ui/drawer/UiDrawer.vue'

defineProps<{
  modelValue: boolean
  title: string
  subtitle?: string
  maxWidthClass?: string
}>()
</script>
