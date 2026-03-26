<template>
  <Teleport to="body">
    <Transition name="ui-drawer" :duration="{ enter: 360, leave: 340 }">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex justify-end"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
      >
        <div class="ui-drawer__backdrop absolute inset-0 z-0 bg-slate-900/40" aria-hidden="true" />
        <aside
          :class="[
            'ui-drawer__panel relative z-10 flex h-full max-h-[100dvh] min-h-0 w-full flex-col bg-white shadow-2xl will-change-transform',
            maxWidthClass,
          ]"
        >
          <header class="flex shrink-0 items-start justify-between gap-3 border-b border-slate-200 px-5 py-4">
            <div class="min-w-0">
              <h2 :id="titleId" class="text-base font-semibold text-slate-900">{{ title }}</h2>
              <p v-if="subtitle" class="mt-1 text-sm text-slate-500">{{ subtitle }}</p>
            </div>
          </header>
          <div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-5 py-4">
            <slot />
          </div>
          <footer
            v-if="$slots.footer"
            class="shrink-0 border-t border-slate-200 bg-white px-5 py-4 shadow-[0_-4px_12px_-4px_rgba(15,23,42,0.08)]"
          >
            <slot name="footer" />
          </footer>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onUnmounted, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    subtitle?: string
    maxWidthClass?: string
  }>(),
  { maxWidthClass: 'max-w-xl' },
)

const titleId = `ui-drawer-title-${Math.random().toString(36).slice(2, 9)}`
const maxWidthClass = computed(() => props.maxWidthClass)

watch(
  () => props.modelValue,
  (v) => {
    if (typeof document === 'undefined') return
    document.body.style.overflow = v ? 'hidden' : ''
  },
  { immediate: true },
)

onUnmounted(() => {
  if (typeof document !== 'undefined') document.body.style.overflow = ''
})
</script>

<style scoped>
.ui-drawer-enter-active,
.ui-drawer-leave-active {
  transition: none;
}

.ui-drawer-enter-active .ui-drawer__backdrop,
.ui-drawer-leave-active .ui-drawer__backdrop {
  transition: opacity 0.28s cubic-bezier(0.4, 0, 0.2, 1);
}

.ui-drawer-enter-active .ui-drawer__panel,
.ui-drawer-leave-active .ui-drawer__panel {
  transition: transform 0.32s cubic-bezier(0.32, 0.72, 0, 1);
}

.ui-drawer-enter-from .ui-drawer__backdrop {
  opacity: 0;
}
.ui-drawer-enter-from .ui-drawer__panel {
  transform: translate3d(100%, 0, 0);
}

.ui-drawer-leave-to .ui-drawer__backdrop {
  opacity: 0;
}
.ui-drawer-leave-to .ui-drawer__panel {
  transform: translate3d(100%, 0, 0);
}
</style>
