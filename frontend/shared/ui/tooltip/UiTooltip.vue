<template>
  <span
    ref="rootRef"
    class="relative inline-flex"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
    @click="onClickTrigger"
  >
    <slot />
    <Transition name="ui-tooltip-fade">
      <span
        v-show="visible"
        class="pointer-events-none absolute z-[450] rounded-md bg-slate-900 px-2 py-1 text-xs text-white shadow-lg"
        :class="placementClass"
        :style="{ maxWidth }"
        role="tooltip"
      >
        <slot name="content">{{ content }}</slot>
      </span>
    </Transition>
  </span>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'

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

const visible = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const placementClass = computed(() => {
  if (props.placement === 'bottom') return 'left-1/2 top-full mt-2 -translate-x-1/2'
  if (props.placement === 'left') return 'right-full top-1/2 mr-2 -translate-y-1/2'
  if (props.placement === 'right') return 'left-full top-1/2 ml-2 -translate-y-1/2'
  return 'bottom-full left-1/2 mb-2 -translate-x-1/2'
})

function onMouseEnter() {
  if (props.trigger === 'hover') visible.value = true
}

function onMouseLeave() {
  if (props.trigger === 'hover') visible.value = false
}

function onClickOutside(e: MouseEvent) {
  const el = rootRef.value
  if (!el) return
  if (!el.contains(e.target as Node)) visible.value = false
}

function onClickTrigger() {
  if (props.trigger !== 'click') return
  visible.value = !visible.value
}

if (typeof document !== 'undefined') {
  document.addEventListener('click', onClickOutside)
}

onBeforeUnmount(() => {
  document.removeEventListener('click', onClickOutside)
})
</script>

<style scoped>
.ui-tooltip-fade-enter-active,
.ui-tooltip-fade-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.ui-tooltip-fade-enter-from,
.ui-tooltip-fade-leave-to {
  opacity: 0;
}
</style>
