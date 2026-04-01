<script setup lang="ts">
import type { TocLink } from '@nuxt/content'

const { data: navigation } = await useAsyncData(
  'navigation',
  () => queryCollectionNavigation('content')
)

const links = computed(() => navigation.value ?? [])

// Read the TOC that the current page publishes via useState.
// Defaults to an empty array so the sidebar renders nothing until a page loads.
const tocLinks = useState<TocLink[]>('docs-toc', () => [])
</script>

<template>
  <UContainer class="py-8">
    <div class="flex gap-8">
      <!-- Left sidebar: category navigation tree (~240px) -->
      <aside class="hidden lg:block w-60 shrink-0">
        <nav class="sticky top-8 max-h-[calc(100vh-4rem)] overflow-y-auto">
          <UContentNavigation
            :navigation="links"
            highlight
          />
        </nav>
      </aside>

      <!-- Center: page content -->
      <main class="min-w-0 flex-1">
        <slot />
      </main>

      <!-- Right sidebar: in-page table of contents (~200px) -->
      <aside
        v-if="tocLinks.length"
        class="hidden xl:block w-52 shrink-0"
      >
        <div class="sticky top-8">
          <UContentToc
            :links="tocLinks"
            highlight
          />
        </div>
      </aside>
    </div>
  </UContainer>
</template>
