<script setup lang="ts">
defineProps<{
  headline: string
  subline: string
  ctaLabel: string
  ctaHref: string
}>()
</script>

<template>
  <section class="flex flex-col items-center text-center px-6 py-24 gap-8">
    <img
      src="~/assets/images/mobile-logo.png"
      alt="Forgia"
      class="h-16 w-auto"
    >
    <div class="flex flex-col items-center gap-4 max-w-3xl">
      <h1 class="text-4xl md:text-6xl font-bold text-highlighted">
        {{ headline }}
      </h1>
      <p class="text-lg md:text-xl text-muted max-w-xl">
        {{ subline }}
      </p>
    </div>
    <UButton
      :label="ctaLabel"
      :to="ctaHref"
      color="primary"
      size="xl"
    />
  </section>
</template>
