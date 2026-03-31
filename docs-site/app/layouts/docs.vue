<script setup lang="ts">
const { navDirFromPath } = useContentHelpers()
const { data: navigation } = await useAsyncData('navigation', () => queryCollectionNavigation('content'))

const links = computed(() => navigation.value ?? [])
</script>

<template>
  <UContainer class="py-8">
    <div class="flex gap-8">
      <!-- Sidebar -->
      <aside class="hidden lg:block w-56 shrink-0">
        <nav class="sticky top-8">
          <UNavigationMenu
            :items="links"
            orientation="vertical"
            class="w-full"
          />
        </nav>
      </aside>

      <!-- Main content -->
      <main class="min-w-0 flex-1">
        <slot />
      </main>
    </div>
  </UContainer>
</template>
