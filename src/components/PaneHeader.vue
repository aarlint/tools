<template>
  <div class="pane-header">
    <span class="pane-label">{{ label }}</span>
    <span class="pane-meta"><slot /></span>
  </div>
</template>

<script setup>
defineProps({
  label: String,
})
</script>

<style>
.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 32px;
  min-height: 32px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-surface);
}

.pane-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.pane-meta {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  color: var(--text-muted);
}
</style>
