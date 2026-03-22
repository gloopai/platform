<template>
  <div
    v-if="message"
    class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800"
    role="alert"
  >
    {{ message }}
  </div>
</template>

<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>
