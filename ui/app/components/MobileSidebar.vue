<script setup lang="ts">
import { mobileDrawerVisible, closeMobileDrawer } from "@/composables/useLayout";

const config = useRuntimeConfig()
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

function isActive(path: string) {
  return route.path === path;
}
</script>

<template>
  <Drawer
    v-model:visible="mobileDrawerVisible"
    position="left"
    class="w-72"
    @hide="closeMobileDrawer"
    :header="config.public.appName"
  >
    <nav class="space-y-1 mt-2">
      <RouterLink
        v-for="item in menuItems"
        :key="item.to"
        :to="item.to"
        @click="closeMobileDrawer"
        class="flex items-center gap-3 px-4 py-3 rounded-md transition"
        :class="isActive(item.to)
          ? 'bg-primary text-primary-contrast'
          : 'hover:bg-primary/10 hover:text-primary'"
      >
        <i :class="item.icon" class="text-lg" />
        <span class="text-sm font-medium">{{ item.label }}</span>
      </RouterLink>
    </nav>
    
  </Drawer>
</template>
