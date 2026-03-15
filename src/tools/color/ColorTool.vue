<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <input type="color" v-model="hex" class="color-native-picker">
      <div class="spacer"></div>
      <ShareButton tool="color" :getState="() => ({ hex })" />
      <button class="toolbar-btn" @click="randomColor">Random</button>
    </div>
    <div class="tool-scroll">
      <div class="color-preview" :style="{ background: hex }">
        <span class="color-preview-label" :style="{ color: isDark ? '#fff' : '#000' }">{{ hex }}</span>
      </div>

      <div class="color-inputs">
        <div class="color-input-group">
          <label class="color-label">HEX</label>
          <div class="color-input-row">
            <input type="text" v-model="hex" spellcheck="false" class="color-field">
            <button class="toolbar-btn" @click="copy(hex, 'HEX copied')">Copy</button>
          </div>
        </div>
        <div class="color-input-group">
          <label class="color-label">RGB</label>
          <div class="color-input-row">
            <input type="number" v-model.number="r" min="0" max="255" class="color-field-sm">
            <input type="number" v-model.number="g" min="0" max="255" class="color-field-sm">
            <input type="number" v-model.number="b" min="0" max="255" class="color-field-sm">
            <button class="toolbar-btn" @click="copy(rgbStr, 'RGB copied')">Copy</button>
          </div>
        </div>
        <div class="color-input-group">
          <label class="color-label">HSL</label>
          <div class="color-input-row">
            <input type="number" v-model.number="h" min="0" max="360" class="color-field-sm">
            <input type="number" v-model.number="s" min="0" max="100" class="color-field-sm">
            <input type="number" v-model.number="l" min="0" max="100" class="color-field-sm">
            <button class="toolbar-btn" @click="copy(hslStr, 'HSL copied')">Copy</button>
          </div>
        </div>
      </div>

      <div class="color-info">
        <div class="kv-row">
          <span class="kv-label">Luminance</span>
          <span class="kv-value">{{ luminance.toFixed(3) }}</span>
        </div>
        <div class="kv-row">
          <span class="kv-label">Appearance</span>
          <span class="kv-value">{{ isDark ? 'Dark' : 'Light' }}</span>
        </div>
        <div class="kv-row">
          <span class="kv-label">CSS</span>
          <span class="kv-value" style="cursor: pointer" @click="copy(`color: ${hex};`, 'CSS copied')">color: {{ hex }};</span>
        </div>
      </div>

      <div class="color-palette-section">
        <div class="color-label">Palette</div>
        <div class="color-palette">
          <div
            v-for="(c, i) in palette"
            :key="i"
            class="color-swatch"
            :style="{ background: c }"
            :title="c"
            @click="copy(c, 'Copied ' + c)"
          >
            <span class="swatch-label" :style="{ color: swatchTextColor(c) }">{{ c }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const { useShareable } = useShare()

const hex = ref('#c4956a')

useShareable('color', (s) => {
  if (s.hex) hex.value = s.hex
})
const r = ref(196)
const g = ref(149)
const b = ref(106)
const h = ref(29)
const s = ref(45)
const l = ref(59)

let updating = false

function hexToRgb(hex) {
  const m = hex.replace('#', '').match(/.{2}/g)
  if (!m || m.length < 3) return null
  return { r: parseInt(m[0], 16), g: parseInt(m[1], 16), b: parseInt(m[2], 16) }
}

function rgbToHex(r, g, b) {
  return '#' + [r, g, b].map(v => Math.max(0, Math.min(255, Math.round(v))).toString(16).padStart(2, '0')).join('')
}

function rgbToHsl(r, g, b) {
  r /= 255; g /= 255; b /= 255
  const max = Math.max(r, g, b), min = Math.min(r, g, b)
  let h = 0, s = 0, l = (max + min) / 2
  if (max !== min) {
    const d = max - min
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
    switch (max) {
      case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break
      case g: h = ((b - r) / d + 2) / 6; break
      case b: h = ((r - g) / d + 4) / 6; break
    }
  }
  return { h: Math.round(h * 360), s: Math.round(s * 100), l: Math.round(l * 100) }
}

function hslToRgb(h, s, l) {
  h /= 360; s /= 100; l /= 100
  let r, g, b
  if (s === 0) { r = g = b = l }
  else {
    const hue2rgb = (p, q, t) => {
      if (t < 0) t += 1; if (t > 1) t -= 1
      if (t < 1/6) return p + (q - p) * 6 * t
      if (t < 1/2) return q
      if (t < 2/3) return p + (q - p) * (2/3 - t) * 6
      return p
    }
    const q = l < 0.5 ? l * (1 + s) : l + s - l * s
    const p = 2 * l - q
    r = hue2rgb(p, q, h + 1/3)
    g = hue2rgb(p, q, h)
    b = hue2rgb(p, q, h - 1/3)
  }
  return { r: Math.round(r * 255), g: Math.round(g * 255), b: Math.round(b * 255) }
}

watch(hex, (val) => {
  if (updating) return
  const rgb = hexToRgb(val)
  if (!rgb) return
  updating = true
  r.value = rgb.r; g.value = rgb.g; b.value = rgb.b
  const hsl = rgbToHsl(rgb.r, rgb.g, rgb.b)
  h.value = hsl.h; s.value = hsl.s; l.value = hsl.l
  updating = false
})

watch([r, g, b], () => {
  if (updating) return
  updating = true
  hex.value = rgbToHex(r.value, g.value, b.value)
  const hsl = rgbToHsl(r.value, g.value, b.value)
  h.value = hsl.h; s.value = hsl.s; l.value = hsl.l
  updating = false
})

watch([h, s, l], () => {
  if (updating) return
  updating = true
  const rgb = hslToRgb(h.value, s.value, l.value)
  r.value = rgb.r; g.value = rgb.g; b.value = rgb.b
  hex.value = rgbToHex(rgb.r, rgb.g, rgb.b)
  updating = false
})

const rgbStr = computed(() => `rgb(${r.value}, ${g.value}, ${b.value})`)
const hslStr = computed(() => `hsl(${h.value}, ${s.value}%, ${l.value}%)`)

const luminance = computed(() => {
  const srgb = [r.value, g.value, b.value].map(v => {
    v /= 255
    return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)
  })
  return 0.2126 * srgb[0] + 0.7152 * srgb[1] + 0.0722 * srgb[2]
})

const isDark = computed(() => luminance.value < 0.5)

const palette = computed(() => {
  const colors = []
  for (let i = 90; i >= 10; i -= 10) {
    colors.push(`hsl(${h.value}, ${s.value}%, ${i}%)`)
  }
  return colors
})

function swatchTextColor(color) {
  const match = color.match(/hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/)
  if (match && parseInt(match[3]) > 55) return '#111'
  return '#eee'
}

function randomColor() {
  hex.value = '#' + Math.floor(Math.random() * 16777215).toString(16).padStart(6, '0')
}
</script>

<style scoped>
.color-native-picker {
  width: 32px;
  height: 28px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: transparent;
  cursor: pointer;
  padding: 0;
}

.color-preview {
  height: 120px;
  border-radius: var(--radius);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  border: 1px solid var(--border);
}

.color-preview-label {
  font-family: var(--font-mono);
  font-size: 1.2rem;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.color-inputs {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
}

.color-input-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.color-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.color-input-row {
  display: flex;
  gap: 6px;
  align-items: center;
}

.color-field {
  flex: 1;
  max-width: 200px;
}

.color-field-sm {
  width: 70px;
}

.color-info {
  margin-bottom: 24px;
}

.color-palette-section {
  margin-bottom: 24px;
}

.color-palette {
  display: flex;
  gap: 0;
  margin-top: 8px;
  border-radius: var(--radius);
  overflow: hidden;
}

.color-swatch {
  flex: 1;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform var(--transition);
}

.color-swatch:hover {
  transform: scaleY(1.15);
  z-index: 1;
}

.swatch-label {
  font-family: var(--font-mono);
  font-size: 0.55rem;
  opacity: 0;
  transition: opacity var(--transition);
}

.color-swatch:hover .swatch-label {
  opacity: 1;
}
</style>
