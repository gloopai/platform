<template>
  <div>
    <h1 class="text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl">{{ title }}</h1>
    <p v-if="description" class="mt-1 text-sm text-slate-600">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>
