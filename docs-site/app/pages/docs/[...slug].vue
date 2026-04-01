<script setup lang="ts">
import type { TocLink } from '@nuxt/content'

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

// Share the TOC with the layout via shared state keyed by the current route path.
// The layout reads this same key to render the right-hand table of contents.
const tocLinks = useState<TocLink[]>('docs-toc', () => page.value?.body?.toc?.links ?? [])

// Keep the shared state in sync when navigating between pages on the client.
watch(page, (newPage) => {
  tocLinks.value = newPage?.body?.toc?.links ?? []
}, { immediate: true })

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
