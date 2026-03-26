<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fixed inset-0 z-[500] bg-slate-900/30" @click="$emit('update:modelValue', false)" />
    <aside
      v-if="modelValue"
      class="fixed inset-y-0 right-0 z-[510] flex w-full max-w-[560px] flex-col border-l border-slate-200 bg-white shadow-2xl"
      @click.stop
    >
        <header class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div class="text-sm font-semibold text-slate-900">{{ title }}</div>
          <button
            type="button"
            class="rounded border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-600 hover:bg-slate-50"
            @click="$emit('update:modelValue', false)"
          >
            关闭
          </button>
        </header>
        <div class="min-h-0 flex-1 overflow-y-auto p-4">
          <slot />
        </div>
        <footer v-if="$slots.footer" class="border-t border-slate-100 p-4">
          <slot name="footer" />
        </footer>
    </aside>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean
  title: string
}>()

defineEmits<{
  (e: 'update:modelValue', v: boolean): void
}>()
</script>
