<template>
  <Teleport to="body">
    <div class="toast" :class="{ show: visible }">{{ message }}</div>
  </Teleport>
</template>

<script setup>
import { useToast } from '../composables/useToast'

const { message, visible } = useToast()
</script>

<style>
.toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%) translateY(20px);
  font-family: var(--font-sans);
  font-size: 0.8rem;
  font-weight: 500;
  padding: 8px 20px;
  background: var(--bg-raised);
  color: var(--accent);
  border: 1px solid var(--accent-dim);
  border-radius: var(--radius);
  opacity: 0;
  transition: all 300ms ease;
  pointer-events: none;
  z-index: 100;
}

.toast.show {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}
</style>
