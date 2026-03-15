<template>
  <nav class="sidebar">
    <div class="sidebar-brand">
      <h1>tools.</h1>
    </div>
    <div class="sidebar-nav">
      <template v-for="group in groups" :key="group.name">
        <div class="sidebar-group-label">{{ group.name }}</div>
        <router-link
          v-for="route in group.routes"
          :key="route.path"
          :to="route.path"
          class="sidebar-link"
          :class="{ active: $route.path === route.path }"
          @click="$emit('navigate')"
        >
          <span class="sidebar-icon">{{ route.meta.icon }}</span>
          <span class="sidebar-name">{{ route.meta.name }}</span>
        </router-link>
      </template>
    </div>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { routes } from '../router'

defineEmits(['navigate'])

const toolRoutes = routes.filter(r => r.meta)

const groupOrder = ['Text', 'Data', 'Encode', 'Dev']
const groups = computed(() => {
  return groupOrder.map(name => ({
    name,
    routes: toolRoutes.filter(r => r.meta.group === name),
  })).filter(g => g.routes.length)
})
</script>

<style>
.sidebar {
  width: var(--sidebar-width);
  min-width: var(--sidebar-width);
  background: var(--bg-surface);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  user-select: none;
}

.sidebar-brand {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.sidebar-brand h1 {
  font-family: var(--font-serif);
  font-size: 1.4rem;
  font-weight: 400;
  color: var(--accent);
  letter-spacing: -0.02em;
}

.sidebar-nav {
  padding: 8px 0;
  flex: 1;
}

.sidebar-group-label {
  font-size: 0.6rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  padding: 14px 20px 6px;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 20px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.82rem;
  font-weight: 400;
  transition: all var(--transition);
  border-left: 2px solid transparent;
}

.sidebar-link:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.sidebar-link.active {
  color: var(--accent);
  background: var(--accent-glow);
  border-left-color: var(--accent);
}

.sidebar-icon {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  width: 22px;
  text-align: center;
  flex-shrink: 0;
  color: var(--text-muted);
}

.sidebar-link.active .sidebar-icon {
  color: var(--accent-dim);
}

.sidebar-name {
  flex: 1;
}
</style>
