<script setup lang="ts">
interface Feature {
  icon: string
  title: string
  description: string
}

defineProps<{
  features: Feature[]
  heading?: string
}>()
</script>

<template>
  <section class="px-6 py-16">
    <div class="max-w-5xl mx-auto flex flex-col gap-10">
      <h2
        v-if="heading"
        class="text-3xl font-bold text-center text-highlighted"
      >
        {{ heading }}
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <UCard
          v-for="feature in features"
          :key="feature.title"
        >
          <div class="flex flex-col gap-3">
            <UIcon
              :name="feature.icon"
              class="w-8 h-8 text-primary"
            />
            <h3 class="font-semibold text-lg text-highlighted">
              {{ feature.title }}
            </h3>
            <p class="text-muted text-sm leading-relaxed">
              {{ feature.description }}
            </p>
          </div>
        </UCard>
      </div>
    </div>
  </section>
</template>
