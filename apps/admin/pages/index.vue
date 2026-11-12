<script setup lang="ts">
import { useApiClient } from '~/shared/api/client';
import type { AppUser } from '~/shared/api/types';

definePageMeta({ layout: 'default' });

const api = useApiClient();
const usersCount = ref<number | null>(null);

onMounted(async () => {
  try {
    const users = await api.get<AppUser[]>('/users');
    usersCount.value = users.length;
  } catch {
    usersCount.value = null;
  }
});
</script>

<template>
  <div class="max-w-4xl">
    <h1 class="mb-1 text-xl font-semibold">Главная</h1>
    <p class="mb-6 text-sm text-steel-500 dark:text-steel-400">
      Стартовая страница админки. Добавляйте разделы в <code>pages/</code> и пункты меню в
      <code>layouts/default.vue</code>.
    </p>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <UCard>
        <div class="flex items-center gap-3">
          <span
            class="flex size-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600 dark:bg-brand-950 dark:text-brand-400"
          >
            <UIcon name="i-lucide-users" class="size-5" />
          </span>
          <div>
            <p class="text-sm text-steel-500 dark:text-steel-400">Пользователи</p>
            <p class="text-2xl font-semibold">{{ usersCount ?? '—' }}</p>
          </div>
        </div>
        <template #footer>
          <UButton to="/users" variant="link" color="primary" trailing-icon="i-lucide-arrow-right" class="p-0">
            Открыть
          </UButton>
        </template>
      </UCard>
    </div>
  </div>
</template>
