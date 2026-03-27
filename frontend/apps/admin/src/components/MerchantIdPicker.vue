<template>
  <div ref="rootEl" class="relative w-full">
    <label class="grid gap-1 text-xs font-medium text-slate-600">
      <span>{{ label }}</span>
      <div class="flex gap-2">
        <input
          v-model="localText"
          type="text"
          autocomplete="off"
          spellcheck="false"
          class="min-w-0 flex-1 rounded-lg border border-slate-200 bg-white px-3 py-2 font-mono text-sm text-slate-900"
          :placeholder="placeholder"
          @focus="open = true"
          @blur="onBlur"
          @keydown.escape="open = false"
          @keydown.enter.prevent="onEnter"
        />
        <button
          v-if="modelValue.trim()"
          type="button"
          class="shrink-0 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm hover:bg-slate-50"
          title="清除"
          @click="clear"
        >
          清除
        </button>
      </div>
    </label>

    <div
      v-show="open && filtered.length > 0"
      class="absolute z-40 mt-1 flex max-h-60 w-full flex-col overflow-hidden rounded-lg border border-slate-200 bg-white shadow-lg"
    >
      <div
        v-if="truncatedHint"
        class="shrink-0 border-b border-slate-100 bg-slate-50/90 px-3 py-1.5 text-[11px] leading-snug text-slate-500"
      >
        {{ truncatedHint }}
      </div>
      <div class="max-h-52 overflow-y-auto py-1">
        <button
          v-for="m in pickable"
          :key="m.merchant_id"
          type="button"
          class="flex w-full items-center px-3 py-2 text-left text-sm text-slate-800 hover:bg-slate-50"
          @mousedown.prevent="choose(m.merchant_id)"
        >
          <span class="font-mono text-sm">{{ m.merchant_id }}</span>
        </button>
      </div>
    </div>

    <p v-if="open && filtered.length === 0 && qnorm" class="mt-1 text-xs text-slate-500">
      无匹配项；可回车使用当前输入的商户 ID。
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

export type MerchantPickRow = { merchant_id: string }

const props = withDefaults(
  defineProps<{
    modelValue: string
    merchants: MerchantPickRow[]
    label?: string
    placeholder?: string
  }>(),
  {
    label: '商户 ID',
    placeholder: '输入关键字筛选，或从列表选择',
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: string]
}>()

const localText = ref(props.modelValue || '')
const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)

watch(
  () => props.modelValue,
  (v) => {
    localText.value = v || ''
  },
)

/** 下拉最多渲染条数，避免商户极多时 DOM 过长、卡顿 */
const MAX_DISPLAY = 80

const qnorm = computed(() => localText.value.trim().toLowerCase())

const filtered = computed(() => {
  const list = props.merchants || []
  const q = qnorm.value
  if (!q) return list
  return list.filter((m) => m.merchant_id.toLowerCase().includes(q))
})

const pickable = computed(() => filtered.value.slice(0, MAX_DISPLAY))

const truncatedHint = computed(() => {
  const n = filtered.value.length
  if (n <= MAX_DISPLAY) return ''
  if (!qnorm.value) {
    return `共 ${n} 个商户，此处最多展示 ${MAX_DISPLAY} 条；请输入关键字筛选后再选。`
  }
  return `共 ${n} 条匹配，此处最多展示 ${MAX_DISPLAY} 条；请继续输入以缩小范围。`
})

function choose(id: string) {
  localText.value = id
  emit('update:modelValue', id.trim())
  open.value = false
}

function clear() {
  localText.value = ''
  emit('update:modelValue', '')
}

function onEnter() {
  const t = localText.value.trim()
  if (filtered.value.length === 1) {
    choose(filtered.value[0].merchant_id)
    return
  }
  emit('update:modelValue', t)
  open.value = false
}

function onBlur() {
  window.setTimeout(() => {
    emit('update:modelValue', localText.value.trim())
    open.value = false
  }, 150)
}

function onDocClick(e: MouseEvent) {
  const el = rootEl.value
  if (!el || !(e.target instanceof Node)) return
  if (!el.contains(e.target)) open.value = false
}

onMounted(() => document.addEventListener('click', onDocClick, true))
onUnmounted(() => document.removeEventListener('click', onDocClick, true))
</script>
