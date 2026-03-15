<template>
  <div class="drop-overlay" :class="{ active: active }">
    <div class="drop-zone">
      <div class="drop-zone-icon">&#8615;</div>
      <div class="drop-zone-text">{{ text || 'Drop your file' }}</div>
      <div class="drop-zone-hint">{{ hint }}</div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  active: Boolean,
  text: String,
  hint: String,
})
</script>

<style>
.drop-overlay {
  position: fixed;
  inset: 0;
  background: rgba(10, 10, 12, 0.92);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  opacity: 0;
  pointer-events: none;
  transition: opacity 250ms ease;
  backdrop-filter: blur(8px);
}

.drop-overlay.active {
  opacity: 1;
  pointer-events: auto;
}

.drop-zone {
  width: 420px;
  height: 280px;
  border: 2px dashed var(--accent-dim);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  transition: border-color var(--transition), background var(--transition);
  background: var(--accent-glow);
}

.drop-zone-icon {
  font-size: 3rem;
  color: var(--accent);
  opacity: 0.6;
  line-height: 1;
}

.drop-zone-text {
  font-family: var(--font-serif);
  font-size: 1.3rem;
  color: var(--text-secondary);
  font-style: italic;
}

.drop-zone-hint {
  font-size: 0.72rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}
</style>
