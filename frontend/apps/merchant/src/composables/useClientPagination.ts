import { computed, ref, watch, type Ref } from 'vue'

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
