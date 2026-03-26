<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-[2147483000] flex items-center justify-center bg-slate-950/45 p-4"
      @click.self="onBackdropClick"
    >
      <section
        :class="['w-full rounded-2xl border border-slate-200 bg-white shadow-2xl pointer-events-auto', maxWidthClass]"
        role="dialog"
        aria-modal="true"
        :aria-label="title"
        @click.stop
      >
        <header v-if="title || $slots.header" class="border-b border-slate-200 px-5 py-4">
          <slot name="header">
            <div class="text-base font-semibold text-slate-900">{{ title }}</div>
          </slot>
        </header>
        <div class="px-5 py-4">
          <slot />
        </div>
        <footer v-if="$slots.footer" class="border-t border-slate-200 px-5 py-4">
          <slot name="footer" />
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    maxWidthClass?: string
    closeOnBackdrop?: boolean
  }>(),
  {
    title: '',
    maxWidthClass: 'max-w-md',
    closeOnBackdrop: true,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
}>()

function onBackdropClick() {
  if (!props.closeOnBackdrop) return
  emit('update:modelValue', false)
}
</script>
