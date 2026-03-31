<script setup lang="ts">
definePageMeta({
  layout: 'docs'
})

const route = useRoute()
const slug = computed(() => {
  const s = route.params.slug
  return Array.isArray(s) ? s.join('/') : (s ?? '')
})

const { data: page } = await useAsyncData(
  `docs-${slug.value}`,
  () => queryCollection('content').path(`/docs/${slug.value}`).first()
)

if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Page not found' })
}

useSeoMeta({
  title: page.value.title,
  description: page.value.description
})
</script>

<template>
  <div>
    <ContentRenderer
      v-if="page"
      :value="page"
      class="prose prose-neutral dark:prose-invert max-w-none"
    />
  </div>
</template>
