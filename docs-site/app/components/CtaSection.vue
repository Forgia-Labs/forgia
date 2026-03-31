<script setup lang="ts">
defineProps<{
  label: string
  href: string
  secondaryLabel?: string
  secondaryHref?: string
  heading?: string
}>()
</script>

<template>
  <section class="px-6 py-16 bg-primary/10">
    <div class="max-w-2xl mx-auto flex flex-col items-center text-center gap-6">
      <p
        v-if="heading"
        class="text-2xl font-bold text-highlighted"
      >
        {{ heading }}
      </p>
      <div class="flex flex-wrap gap-4 justify-center">
        <UButton
          :label="label"
          :to="href"
          color="primary"
          size="lg"
          :external="href.startsWith('http')"
          :target="href.startsWith('http') ? '_blank' : undefined"
          :rel="href.startsWith('http') ? 'noopener noreferrer' : undefined"
        />
        <UButton
          v-if="secondaryLabel && secondaryHref"
          :label="secondaryLabel"
          :to="secondaryHref"
          color="neutral"
          variant="outline"
          size="lg"
          :external="secondaryHref.startsWith('http')"
          :target="secondaryHref.startsWith('http') ? '_blank' : undefined"
          :rel="secondaryHref.startsWith('http') ? 'noopener noreferrer' : undefined"
        />
      </div>
    </div>
  </section>
</template>
