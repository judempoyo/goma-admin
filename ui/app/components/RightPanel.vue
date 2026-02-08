<script setup lang="ts">
import { sidebarOpen } from "@/composables/useLayout";

const route = useRoute();

const menuItems = [
  { label: "Dashboard", icon: "pi pi-home", to: "/" },
  { label: "Routes", icon: "pi pi-compass", to: "/routes" },
  { label: "Middlewares", icon: "pi pi-filter", to: "/middlewares" },
  { label: "Instances", icon: "pi pi-server", to: "/instances" },
  { label: "Analytics", icon: "pi pi-chart-line", to: "/analytics" },
  { label: "Configuration", icon: "pi pi-cog", to: "/configuration" },
  { label: "History", icon: "pi pi-history", to: "/history" },
];

const sidebarWidth = computed(() =>
  sidebarOpen.value ? "w-56" : "w-16"
);

function isActive(path: string) {
  return route.path === path;
}
</script>

<template>
  <aside
    class="fixed top-10 left-0 z-40 h-[calc(100vh-2.5rem)]
           bg-surface-0 dark:bg-surface-900
           border-r border-surface-200 dark:border-surface-800
           transition-all duration-300 ease-in-out
           hidden lg:flex flex-col"
    :class="sidebarWidth"
  >
    <nav class="flex-1 px-2 py-3 space-y-1">
      <RouterLink
        v-for="item in menuItems"
        :key="item.to"
        :to="item.to"
        v-tooltip="!sidebarOpen ? item.label : undefined"
        class="group relative flex items-center gap-3 px-3 py-2 rounded-md
               transition-all duration-200"
        :class="[
          isActive(item.to)
            ? 'bg-primary text-primary-contrast shadow-sm'
            : 'text-surface-600 dark:text-surface-300',
          !isActive(item.to)
            ? 'hover:bg-primary/10 dark:hover:bg-primary/20 hover:text-primary'
            : ''
        ]"
      >
        <span
          v-if="isActive(item.to)"
          class="absolute left-0 top-1 bottom-1 w-1 rounded-r bg-primary-contrast"
        />

        <i
          :class="item.icon"
          class="text-lg transition-transform duration-200
                 group-hover:scale-110"
        />

        <span
          v-if="sidebarOpen"
          class="text-sm font-medium whitespace-nowrap"
        >
          {{ item.label }}
        </span>
      </RouterLink>
    </nav>
  </aside>
</template>

