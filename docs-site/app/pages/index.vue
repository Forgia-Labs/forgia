<script setup lang="ts">
import HeroSection from '~/components/HeroSection.vue'
import FeaturesSection from '~/components/FeaturesSection.vue'
import CtaSection from '~/components/CtaSection.vue'

const heroProps = {
  headline: 'Il Framework per Agenti AI',
  subline: 'Un posto dove gli agenti lavorano su task reali.',
  ctaLabel: 'Vedi su GitHub',
  ctaHref: 'https://github.com/forgia-labs/forgia'
}

const featuresData = [
  {
    icon: 'i-heroicons-cpu-chip',
    title: 'Agenti su Task Reali',
    description: 'Gli agenti AI non simulano — completano task concreti in un ambiente strutturato.'
  },
  {
    icon: 'i-heroicons-wrench-screwdriver',
    title: 'Framework Componibile',
    description: 'Costruisci pipeline modulari con interfacce chiare e testabili.'
  },
  {
    icon: 'i-heroicons-bolt',
    title: 'Veloce da Avviare',
    description: 'Da zero a primo agente in pochi minuti, senza configurazione boilerplate.'
  }
]

const ctaProps = {
  heading: 'Pronto a forgiare?',
  label: 'Inizia la Forgiatura',
  href: 'https://github.com/forgia-labs/forgia',
  secondaryLabel: 'Leggi la documentazione',
  secondaryHref: '/docs'
}
</script>

<template>
  <div>
    <HeroSection v-bind="heroProps" />
    <FeaturesSection
      :features="featuresData"
      heading="Perché Forgia?"
    />
    <CtaSection v-bind="ctaProps" />
  </div>
</template>
