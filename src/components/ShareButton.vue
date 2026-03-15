<template>
  <div class="share-wrapper">
    <button class="toolbar-btn" :class="{ active: open }" @click.stop="open = !open">Share</button>
    <div class="share-menu" :class="{ open }" @click.stop>
      <button class="share-item" v-for="opt in options" :key="opt.value" @click="share(opt.value)" :disabled="loading">
        <span class="share-icon">{{ opt.icon }}</span> {{ opt.label }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useShare } from '../composables/useShare'
import { useClipboard } from '../composables/useClipboard'
import { useToast } from '../composables/useToast'

const props = defineProps({
  tool: { type: String, required: true },
  getState: { type: Function, required: true },
})

const { createShare } = useShare()
const { copy } = useClipboard()
const { toast } = useToast()

const open = ref(false)
const loading = ref(false)

const options = [
  { value: 'once', label: 'One-time view', icon: '1x' },
  { value: '15m', label: '15 minutes', icon: '15m' },
  { value: '1h', label: '1 hour', icon: '1h' },
  { value: '6h', label: '6 hours', icon: '6h' },
  { value: '1d', label: '1 day', icon: '1d' },
  { value: '7d', label: '1 week', icon: '7d' },
]

async function share(ttl) {
  loading.value = true
  open.value = false
  try {
    const state = props.getState()
    const result = await createShare(props.tool, state, ttl)
    await navigator.clipboard.writeText(result.url)
    toast('Share link copied')
  } catch (e) {
    console.error('Share failed:', e)
    toast('Share failed')
  } finally {
    loading.value = false
  }
}

function closeMenu() { open.value = false }

onMounted(() => document.addEventListener('click', closeMenu))
onUnmounted(() => document.removeEventListener('click', closeMenu))
</script>

<style scoped>
.share-wrapper {
  position: relative;
}

.share-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 160px;
  background: var(--bg-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 4px 0;
  opacity: 0;
  visibility: hidden;
  transform: translateY(-4px);
  transition: all 150ms ease;
  z-index: 50;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.share-menu.open {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.share-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 14px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-family: var(--font-sans);
  font-size: 0.78rem;
  font-weight: 400;
  cursor: pointer;
  transition: all var(--transition);
  text-align: left;
}

.share-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.share-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.share-icon {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  color: var(--text-muted);
  width: 24px;
  text-align: right;
  flex-shrink: 0;
  letter-spacing: -0.02em;
}
</style>
