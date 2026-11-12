<script setup lang="ts">
const { t } = useI18n();

useSeoMeta({
  title: () => t('meta.title'),
  description: () => t('meta.description'),
  ogTitle: () => t('meta.title'),
  ogDescription: () => t('meta.description'),
});

// hreflang-ссылки на все локали страницы (генерирует @nuxtjs/i18n).
const i18nHead = useLocaleHead();
useHead(() => ({ link: i18nHead.value.link }));

const config = useRuntimeConfig();
const botUrl = config.public.botUrl as string;

const features = [
  { icon: 'i-lucide-server', key: 'backend' },
  { icon: 'i-lucide-layout-dashboard', key: 'frontend' },
  { icon: 'i-lucide-bot', key: 'bot' },
] as const;
</script>

<template>
  <div>
    <!-- Hero -->
    <section class="relative overflow-hidden">
      <div
        class="pointer-events-none absolute inset-0 -z-10 bg-gradient-to-br from-brand-50 via-white to-steel-50 dark:from-steel-950 dark:via-steel-950 dark:to-steel-900"
      />
      <div class="mx-auto w-full max-w-6xl px-4 py-24 text-center">
        <UBadge color="primary" variant="soft" class="mb-4">{{ $t('hero.badge') }}</UBadge>
        <h1 class="mx-auto max-w-3xl text-4xl font-bold leading-tight sm:text-5xl">
          {{ $t('hero.titleLead') }}
          <span class="text-brand-600 dark:text-brand-400">{{ $t('hero.titleAccent') }}</span>
        </h1>
        <p class="mx-auto mt-4 max-w-xl text-lg text-steel-600 dark:text-steel-300">
          {{ $t('hero.subtitle') }}
        </p>
        <div class="mt-8 flex flex-wrap justify-center gap-3">
          <UButton to="#features" size="lg" color="primary" icon="i-lucide-arrow-down">
            {{ $t('hero.ctaMore') }}
          </UButton>
          <UButton :to="botUrl" target="_blank" size="lg" color="neutral" variant="outline" icon="i-lucide-send">
            {{ $t('hero.ctaBot') }}
          </UButton>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section id="features" class="mx-auto w-full max-w-6xl px-4 py-16">
      <h2 class="text-2xl font-bold sm:text-3xl">{{ $t('features.heading') }}</h2>
      <p class="mt-2 text-steel-600 dark:text-steel-300">{{ $t('features.sub') }}</p>
      <div class="mt-8 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <UCard v-for="feature in features" :key="feature.key">
          <span
            class="flex size-11 items-center justify-center rounded-xl bg-brand-50 text-brand-600 dark:bg-brand-950 dark:text-brand-400"
          >
            <UIcon :name="feature.icon" class="size-6" />
          </span>
          <h3 class="mt-4 text-lg font-semibold">{{ $t(`features.${feature.key}.title`) }}</h3>
          <p class="mt-1 text-sm text-steel-600 dark:text-steel-300">{{ $t(`features.${feature.key}.text`) }}</p>
        </UCard>
      </div>
    </section>
  </div>
</template>
