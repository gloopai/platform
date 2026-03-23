import { computed, ref, watch, type Ref } from 'vue'

/**
 * 对已在内存中的列表做前端分页（筛选后再分页时，请在外部把 `source` 换成筛选后的列表）。
 */
export function useClientPagination<T>(source: Ref<T[]>, initialPageSize = 10) {
  const page = ref(1)
  const pageSize = ref(initialPageSize)
  const total = computed(() => source.value.length)
  const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const slice = computed(() => {
    const start = (page.value - 1) * pageSize.value
    return source.value.slice(start, start + pageSize.value)
  })

  watch([source, pageSize], () => {
    if (page.value > pageCount.value) page.value = Math.max(1, pageCount.value)
  })

  return { page, pageSize, total, pageCount, slice }
}
