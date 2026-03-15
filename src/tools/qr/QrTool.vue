<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <select v-model.number="size">
        <option :value="256">256px</option>
        <option :value="512">512px</option>
        <option :value="1024">1024px</option>
      </select>
      <select v-model="errorLevel">
        <option value="L">Low (L)</option>
        <option value="M">Medium (M)</option>
        <option value="Q">Quartile (Q)</option>
        <option value="H">High (H)</option>
      </select>
      <div class="spacer"></div>
      <ShareButton tool="qr" :getState="() => ({ inputText, size, errorLevel, darkColor, lightColor })" />
      <button class="toolbar-btn" @click="downloadQr" :disabled="!qrDataUrl">Download PNG</button>
    </div>
    <div class="tool-scroll">
      <div class="qr-generate-section">
        <div class="qr-section-label">Generate QR Code</div>
        <textarea
          v-model="inputText"
          placeholder="Enter text, URL, or data to encode..."
          class="qr-input"
          rows="3"
        ></textarea>

        <div class="qr-colors">
          <label class="qr-color-group">
            <span class="qr-color-label">Dark</span>
            <input type="color" v-model="darkColor">
          </label>
          <label class="qr-color-group">
            <span class="qr-color-label">Light</span>
            <input type="color" v-model="lightColor">
          </label>
        </div>

        <div class="qr-preview" v-if="qrDataUrl">
          <img :src="qrDataUrl" alt="QR Code" class="qr-image">
        </div>
        <div class="qr-preview qr-empty" v-else-if="!inputText">
          <div class="qr-empty-text">Enter text above to generate a QR code</div>
        </div>
        <div class="qr-preview qr-error-box" v-else-if="generateError">
          <div class="qr-error">{{ generateError }}</div>
        </div>
      </div>

      <div class="qr-reader-section">
        <div class="qr-section-label">Read QR Code</div>
        <div
          class="qr-drop-zone"
          @click="$refs.fileInput.click()"
          @dragover.prevent
          @drop.prevent="onDrop"
        >
          <input ref="fileInput" type="file" accept="image/*" style="display: none" @change="onFileSelect">
          <div class="qr-drop-text">Click or drop an image to decode</div>
          <div class="qr-drop-hint">.png · .jpg · .gif · .webp</div>
        </div>
        <div class="qr-decoded" v-if="decodedText">
          <div class="qr-section-label">Decoded Content</div>
          <div class="qr-decoded-value" @click="copy(decodedText, 'Decoded text copied')">{{ decodedText }}</div>
        </div>
        <div class="qr-decoded-error" v-if="readError">{{ readError }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import QRCode from 'qrcode'
import jsQR from 'jsqr'
import { useToast } from '../../composables/useToast'
import { useClipboard } from '../../composables/useClipboard'
import { useDownload } from '../../composables/useDownload'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { toast } = useToast()
const { copy } = useClipboard()
const { download } = useDownload()

const { useShareable } = useShare()

const inputText = ref('')
const size = ref(512)
const errorLevel = ref('M')
const darkColor = ref('#e4e4e8')
const lightColor = ref('#0a0a0c')

useShareable('qr', (s) => {
  if (s.inputText) inputText.value = s.inputText
  if (s.size) size.value = s.size
  if (s.errorLevel) errorLevel.value = s.errorLevel
  if (s.darkColor) darkColor.value = s.darkColor
  if (s.lightColor) lightColor.value = s.lightColor
})
const qrDataUrl = ref('')
const generateError = ref('')
const decodedText = ref('')
const readError = ref('')
const fileInput = ref(null)

watch([inputText, size, errorLevel, darkColor, lightColor], async () => {
  if (!inputText.value) { qrDataUrl.value = ''; generateError.value = ''; return }
  try {
    qrDataUrl.value = await QRCode.toDataURL(inputText.value, {
      width: size.value,
      errorCorrectionLevel: errorLevel.value,
      color: {
        dark: darkColor.value,
        light: lightColor.value,
      },
      margin: 2,
    })
    generateError.value = ''
  } catch (e) {
    generateError.value = e.message
    qrDataUrl.value = ''
  }
}, { immediate: true })

function downloadQr() {
  if (!qrDataUrl.value) return
  // Convert data URL to blob
  fetch(qrDataUrl.value)
    .then(r => r.blob())
    .then(blob => {
      download(blob, 'qrcode.png', 'image/png')
      toast('QR code downloaded')
    })
}

function decodeImage(file) {
  readError.value = ''
  decodedText.value = ''

  const img = new Image()
  const reader = new FileReader()

  reader.onload = (e) => {
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = img.width
      canvas.height = img.height
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0)
      const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
      const code = jsQR(imageData.data, canvas.width, canvas.height)
      if (code) {
        decodedText.value = code.data
        toast('QR code decoded')
      } else {
        readError.value = 'No QR code found in this image'
      }
    }
    img.src = e.target.result
  }
  reader.readAsDataURL(file)
}

function onFileSelect(e) {
  const file = e.target.files?.[0]
  if (file) decodeImage(file)
}

function onDrop(e) {
  const file = e.dataTransfer?.files?.[0]
  if (file && file.type.startsWith('image/')) decodeImage(file)
}
</script>

<style scoped>
.qr-section-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.qr-generate-section {
  margin-bottom: 32px;
}

.qr-input {
  width: 100%;
  max-width: 500px;
  resize: vertical;
  min-height: 60px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.87rem;
  line-height: 1.6;
  padding: 12px 16px;
  outline: none;
  transition: border-color var(--transition);
}

.qr-input:focus {
  border-color: var(--accent-dim);
}

.qr-colors {
  display: flex;
  gap: 16px;
  margin: 12px 0;
}

.qr-color-group {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.qr-color-label {
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.qr-color-group input[type="color"] {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: transparent;
  cursor: pointer;
  padding: 0;
}

.qr-preview {
  margin-top: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}

.qr-image {
  max-width: 280px;
  max-height: 280px;
  border-radius: 4px;
  image-rendering: pixelated;
}

.qr-empty {
  min-height: 180px;
}

.qr-empty-text {
  color: var(--text-muted);
  font-style: italic;
  font-size: 0.85rem;
}

.qr-error-box {
  min-height: 100px;
}

.qr-error {
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
}

.qr-reader-section {
  margin-bottom: 32px;
}

.qr-drop-zone {
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  padding: 32px;
  text-align: center;
  cursor: pointer;
  transition: border-color var(--transition), background var(--transition);
}

.qr-drop-zone:hover {
  border-color: var(--accent-dim);
  background: var(--accent-glow);
}

.qr-drop-text {
  color: var(--text-secondary);
  font-size: 0.85rem;
  margin-bottom: 4px;
}

.qr-drop-hint {
  font-size: 0.68rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.qr-decoded {
  margin-top: 16px;
}

.qr-decoded-value {
  font-family: var(--font-mono);
  font-size: 0.9rem;
  color: var(--accent);
  padding: 12px 16px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  word-break: break-all;
  transition: border-color var(--transition);
}

.qr-decoded-value:hover {
  border-color: var(--accent-dim);
}

.qr-decoded-error {
  margin-top: 12px;
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
}
</style>
