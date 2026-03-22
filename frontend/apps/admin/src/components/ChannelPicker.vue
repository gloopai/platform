<template>
  <div ref="rootEl" class="relative min-w-[240px] max-w-lg">
    <label class="grid gap-1">
      <span class="text-xs text-slate-500">{{ label }}</span>
      <div class="flex gap-2">
        <input
          v-model="query"
          type="search"
          autocomplete="off"
          class="min-w-0 flex-1 rounded-md border border-slate-200 px-3 py-2 text-sm"
          :placeholder="placeholder"
          @focus="open = true"
          @keydown.escape="open = false"
        />
        <button
          v-if="modelValue > 0"
          type="button"
          class="shrink-0 rounded-md border border-slate-200 px-2 py-2 text-xs font-medium text-slate-600 hover:bg-slate-50"
          title="清除"
          @click="clear"
        >
          清除
        </button>
      </div>
    </label>

    <div
      v-if="modelValue > 0 && selectedLabel"
      class="mt-1 truncate font-mono text-xs text-slate-600"
    >
      已选：{{ selectedLabel }}
    </div>

    <div
      v-show="open"
      class="absolute z-30 mt-1 max-h-60 w-full overflow-y-auto rounded-lg border border-slate-200 bg-white py-1 shadow-lg"
    >
      <div
        v-if="open && !qnorm && searchable.length > 80"
        class="border-b border-slate-100 px-3 py-1.5 text-[11px] leading-snug text-slate-500"
      >
        共 {{ searchable.length }} 条未绑定通道；未输入时仅列出前 80 条，输入关键字可精确查找。
      </div>
      <div v-if="pickable.length === 0" class="px-3 py-4 text-center text-xs text-slate-500">
        <span v-if="searchable.length === 0">当前产品已绑定全部通道，或平台暂无通道数据</span>
        <span v-else>无匹配通道，请换个关键字</span>
      </div>
      <button
        v-for="c in pickable"
        :key="c.id"
        type="button"
        class="flex w-full items-baseline gap-2 px-3 py-2 text-left text-sm hover:bg-slate-50"
        @mousedown.prevent="choose(c.id)"
      >
        <span class="font-mono text-xs text-slate-500">#{{ c.id }}</span>
        <span class="min-w-0 flex-1 truncate font-medium text-slate-900">{{ c.name }}</span>
        <span class="shrink-0 font-mono text-xs text-slate-500">{{ c.pay_type || '—' }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

export type ChannelPickItem = {
  id: number
  name: string
  pay_type: string
}

const props = withDefaults(
  defineProps<{
    channels: ChannelPickItem[]
    modelValue: number
    /** 已在本产品中绑定的通道，不可重复选择 */
    excludeChannelIds?: number[]
    label?: string
    placeholder?: string
  }>(),
  {
    excludeChannelIds: () => [],
    label: '通道',
    placeholder: '搜索 ID、名称、支付类型…',
  },
)

const emit = defineEmits<{
  'update:modelValue': [id: number]
}>()

const query = ref('')
const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)

const excludeSet = computed(() => new Set(props.excludeChannelIds))

const searchable = computed(() =>
  props.channels.filter((c) => !excludeSet.value.has(c.id)),
)

const qnorm = computed(() => query.value.trim().toLowerCase())

const filtered = computed(() => {
  const s = qnorm.value
  const list = searchable.value
  if (!s) return list
  return list.filter((c) => {
    const idStr = String(c.id)
    const name = (c.name || '').toLowerCase()
    const pt = (c.pay_type || '').toLowerCase()
    return idStr.includes(s) || name.includes(s) || pt.includes(s)
  })
})

/** 下拉中展示：无关键字时只显示前 80 条，避免一次渲染过多 */
const pickable = computed(() => {
  const list = filtered.value
  if (qnorm.value) return list
  return list.slice(0, 80)
})

const selectedLabel = computed(() => {
  if (props.modelValue <= 0) return ''
  const c = props.channels.find((x) => x.id === props.modelValue)
  if (!c) return `#${props.modelValue}`
  return `#${c.id} ${c.name} (${c.pay_type || '-'})`
})

watch(
  () => props.modelValue,
  () => {
    query.value = ''
  },
)

function choose(id: number) {
  emit('update:modelValue', id)
  open.value = false
  query.value = ''
}

function clear() {
  emit('update:modelValue', 0)
  query.value = ''
}

function onDocClick(e: MouseEvent) {
  const el = rootEl.value
  if (!el || !(e.target instanceof Node)) return
  if (!el.contains(e.target)) open.value = false
}

onMounted(() => document.addEventListener('click', onDocClick, true))
onUnmounted(() => document.removeEventListener('click', onDocClick, true))
</script>
