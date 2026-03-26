<template>
  <SharedUiTooltip v-bind="props">
    <slot />
    <template #content>
      <slot name="content" />
    </template>
  </SharedUiTooltip>
</template>

<script setup lang="ts">
import SharedUiTooltip from '../../../../../shared/ui/tooltip/UiTooltip.vue'

const props = withDefaults(
  defineProps<{
    content?: string
    placement?: 'top' | 'right' | 'bottom' | 'left'
    trigger?: 'hover' | 'click'
    maxWidth?: string
  }>(),
  {
    content: '',
    placement: 'top',
    trigger: 'hover',
    maxWidth: '16rem',
  },
)
</script>
