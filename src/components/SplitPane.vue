<template>
  <div class="split-pane" ref="container">
    <div class="split-left" :style="{ flex: `0 0 ${leftPercent}%` }">
      <slot name="left" />
    </div>
    <div
      class="split-divider"
      @mousedown.prevent="startDrag"
      @touchstart.prevent="startDragTouch"
    ></div>
    <div class="split-right" :style="{ flex: `0 0 ${100 - leftPercent}%` }">
      <slot name="right" />
    </div>
  </div>
</template>

<script setup>
import { ref, onUnmounted } from 'vue'

const props = defineProps({
  initialSplit: { type: Number, default: 50 },
})

const container = ref(null)
const leftPercent = ref(props.initialSplit)

let dragging = false

function startDrag() {
  dragging = true
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

function startDragTouch(e) {
  dragging = true
  document.addEventListener('touchmove', onDragTouch)
  document.addEventListener('touchend', stopDrag)
}

function onDrag(e) {
  if (!dragging || !container.value) return
  const rect = container.value.getBoundingClientRect()
  const pct = ((e.clientX - rect.left) / rect.width) * 100
  leftPercent.value = Math.max(15, Math.min(85, pct))
}

function onDragTouch(e) {
  if (!dragging || !container.value || !e.touches[0]) return
  const rect = container.value.getBoundingClientRect()
  const pct = ((e.touches[0].clientX - rect.left) / rect.width) * 100
  leftPercent.value = Math.max(15, Math.min(85, pct))
}

function stopDrag() {
  dragging = false
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.removeEventListener('touchmove', onDragTouch)
  document.removeEventListener('touchend', stopDrag)
}

onUnmounted(stopDrag)
</script>

<style>
.split-pane {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.split-left,
.split-right {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.split-divider {
  width: 1px;
  background: var(--border-subtle);
  position: relative;
  cursor: col-resize;
  flex-shrink: 0;
  z-index: 5;
}

.split-divider::after {
  content: '';
  position: absolute;
  top: 0; bottom: 0;
  left: -4px; right: -4px;
}

.split-divider:hover { background: var(--accent-dim); }
</style>
