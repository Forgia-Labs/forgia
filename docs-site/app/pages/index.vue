<script setup lang="ts">
import HeroSection from '~/components/HeroSection.vue'
import FeaturesSection from '~/components/FeaturesSection.vue'
import CtaSection from '~/components/CtaSection.vue'

const heroProps = {
  headline: 'Spec-Driven Development for AI Agents',
  subline: 'Forge feature designs into agent-ready specs. Ship with confidence.',
  ctaLabel: 'Get Started',
  ctaHref: '/docs/getting-started'
}

const featuresData = [
  {
    icon: 'i-heroicons-cpu-chip',
    title: 'Agents on Real Tasks',
    description: 'AI agents don\'t simulate — they complete concrete tasks in a structured, auditable environment.'
  },
  {
    icon: 'i-heroicons-wrench-screwdriver',
    title: 'Composable Framework',
    description: 'Build modular pipelines with clear interfaces and testable execution contracts.'
  },
  {
    icon: 'i-heroicons-bolt',
    title: 'Fast to Start',
    description: 'From zero to first agent in minutes — no boilerplate configuration required.'
  }
]

const ctaProps = {
  heading: 'Ready to forge?',
  label: 'Read the docs',
  href: '/docs',
  secondaryLabel: 'View on GitHub',
  secondaryHref: 'https://github.com/forgia-labs/forgia'
}
</script>

<template>
  <div>
    <HeroSection v-bind="heroProps" />
    <FeaturesSection
      :features="featuresData"
      heading="Why Forgia?"
    />
    <CtaSection v-bind="ctaProps" />
  </div>
</template>
