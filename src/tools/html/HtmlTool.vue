<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <span class="toolbar-filename" :class="{ visible: filename }">{{ filename }}</span>
      <span class="tag">{{ wordCount }} words</span>
      <span class="tag">{{ charCount }} chars</span>
      <div class="spacer"></div>
      <button class="toolbar-btn" :class="{ active: viewMode === 'split' }" @click="viewMode = 'split'">Split</button>
      <button class="toolbar-btn" :class="{ active: viewMode === 'preview' }" @click="viewMode = 'preview'">Preview</button>
      <ShareButton tool="html" :getState="() => ({ source: source })" />
      <div class="export-wrapper">
        <button class="toolbar-btn" :class="{ active: exportOpen }" @click.stop="exportOpen = !exportOpen">Export</button>
        <div class="export-menu" :class="{ open: exportOpen }" @click.stop>
          <button class="export-item" @click="exportPdf"><span class="export-icon">.pdf</span> PDF Document</button>
          <button class="export-item" @click="exportPng"><span class="export-icon">.png</span> PNG Image</button>
          <div class="export-sep"></div>
          <button class="export-item" @click="exportHtml"><span class="export-icon">.html</span> HTML File</button>
          <button class="export-item" @click="copy(source, 'HTML copied'); closeExport()"><span class="export-icon">clipboard</span> Copy HTML</button>
        </div>
      </div>
    </div>
    <div class="tool-content">
      <SplitPane v-if="viewMode === 'split'" :initial-split="45">
        <template #left>
          <div class="pane editor-pane">
            <PaneHeader label="Editor">{{ charCount }} chars</PaneHeader>
            <textarea
              ref="editorEl"
              v-model="source"
              placeholder="Write or paste HTML here..."
              spellcheck="false"
              @keydown.tab.prevent="insertTab"
            ></textarea>
          </div>
        </template>
        <template #right>
          <div class="pane preview-pane">
            <PaneHeader label="Preview">{{ wordCount }} words</PaneHeader>
            <div class="preview-scroll">
              <iframe
                ref="frameEl"
                class="html-frame"
                sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
                :srcdoc="frameDoc"
              ></iframe>
            </div>
          </div>
        </template>
      </SplitPane>
      <div v-else class="pane preview-pane preview-full">
        <PaneHeader label="Preview">{{ wordCount }} words</PaneHeader>
        <div class="preview-scroll">
          <iframe
            ref="frameEl"
            class="html-frame"
            sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
            :srcdoc="frameDoc"
          ></iframe>
        </div>
      </div>
    </div>
    <DropOverlay :active="dragging" text="Drop your .html file" hint=".html · .htm" />
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import html2canvas from 'html2canvas'
import { jsPDF } from 'jspdf'
import SplitPane from '../../components/SplitPane.vue'
import PaneHeader from '../../components/PaneHeader.vue'
import DropOverlay from '../../components/DropOverlay.vue'
import { useToast } from '../../composables/useToast'
import { useClipboard } from '../../composables/useClipboard'
import { useDownload } from '../../composables/useDownload'
import { useFileDrop } from '../../composables/useFileDrop'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { toast } = useToast()
const { copy } = useClipboard()
const { download } = useDownload()

const exportOpen = ref(false)
const viewMode = ref('split')
const filename = ref('')
const source = ref('')
const editorEl = ref(null)
const frameEl = ref(null)

const { dragging } = useFileDrop((content, name) => {
  source.value = content
  filename.value = name
  toast(`Loaded ${name}`)
}, { extensions: ['.html', '.htm'] })

const charCount = computed(() => source.value.length.toLocaleString())
const wordCount = computed(() => {
  const t = source.value.replace(/<[^>]*>/g, ' ').trim()
  return t ? t.split(/\s+/).length.toLocaleString() : '0'
})

// Wrap fragments so partial HTML still renders. Detect full document.
const isFullDoc = computed(() => /<(!doctype|html|head|body)\b/i.test(source.value))
const frameDoc = computed(() => {
  if (isFullDoc.value) return source.value
  return `<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
    html,body{margin:0;padding:24px;background:#111114;color:#e4e4e8;font-family:system-ui,-apple-system,sans-serif;line-height:1.6;}
    a{color:#7ca5d4;}
    pre,code{font-family:ui-monospace,'IBM Plex Mono',monospace;}
    pre{background:#0a0a0c;padding:12px;border-radius:6px;overflow:auto;}
    table{border-collapse:collapse;}
    th,td{border:1px solid #2a2a30;padding:6px 10px;}
    img{max-width:100%;}
  </style></head><body>${source.value}</body></html>`
})

function insertTab() {
  const el = editorEl.value
  if (!el) return
  const start = el.selectionStart
  const end = el.selectionEnd
  source.value = source.value.substring(0, start) + '  ' + source.value.substring(end)
  nextTick(() => {
    el.selectionStart = el.selectionEnd = start + 2
  })
}

const baseName = computed(() => filename.value?.replace(/\.[^.]+$/, '') || 'document')

function closeExport() { exportOpen.value = false }

onMounted(() => {
  document.addEventListener('click', () => { exportOpen.value = false })
})

async function renderToCanvas() {
  // Snapshot iframe body offscreen at fixed width for consistent export
  const doc = frameEl.value?.contentDocument
  if (!doc || !doc.body) throw new Error('Preview not ready')

  const container = document.createElement('div')
  container.style.cssText = `
    position: absolute; left: -9999px; top: 0;
    width: 816px;
    background: #111114;
    padding: 48px 56px;
    color: #e4e4e8;
    font-family: system-ui, -apple-system, sans-serif;
    line-height: 1.6;
  `
  container.innerHTML = doc.body.innerHTML
  // Copy <style> tags from iframe head so styling carries over
  const styles = doc.querySelectorAll('style, link[rel="stylesheet"]')
  styles.forEach(s => container.prepend(s.cloneNode(true)))
  document.body.appendChild(container)

  const containerWidth = container.offsetWidth
  const containerH = container.scrollHeight
  const containerTop = container.getBoundingClientRect().top

  const children = container.children
  const breakPoints = [0]
  for (let i = 0; i < children.length; i++) {
    breakPoints.push(children[i].getBoundingClientRect().bottom - containerTop)
  }
  if (breakPoints[breakPoints.length - 1] < containerH) {
    breakPoints.push(containerH)
  }

  const canvas = await html2canvas(container, {
    backgroundColor: '#111114',
    scale: 2,
    useCORS: true,
    logging: false,
  })

  document.body.removeChild(container)
  return { canvas, containerWidth, breakPoints }
}

async function exportPdf() {
  closeExport()
  toast('Generating PDF...')
  try {
    const { canvas, containerWidth, breakPoints } = await renderToCanvas()

    const pdfPageWidth = 612, pdfPageHeight = 792, margin = 36
    const contentWidth = pdfPageWidth - margin * 2
    const contentHeight = pdfPageHeight - margin * 2
    const cssToPdf = contentWidth / containerWidth
    const maxCssPerPage = contentHeight / cssToPdf
    const canvasScale = canvas.width / containerWidth

    const pages = []
    let pageStart = 0
    for (let i = 1; i < breakPoints.length; i++) {
      const blockBottom = breakPoints[i]
      const blockTop = breakPoints[i - 1]
      const blockH = blockBottom - blockTop
      if (blockH > maxCssPerPage && pageStart === blockTop) {
        let cursor = pageStart
        while (cursor < blockBottom) {
          const end = Math.min(cursor + maxCssPerPage, blockBottom)
          pages.push({ startPx: cursor, endPx: end })
          cursor = end
        }
        pageStart = blockBottom
        continue
      }
      if (blockBottom - pageStart > maxCssPerPage) {
        if (blockTop > pageStart) {
          pages.push({ startPx: pageStart, endPx: blockTop })
          pageStart = blockTop
        }
        if (blockH > maxCssPerPage) { i--; continue }
      }
    }
    const lastBreak = breakPoints[breakPoints.length - 1]
    if (pageStart < lastBreak) pages.push({ startPx: pageStart, endPx: lastBreak })

    const pdf = new jsPDF({ unit: 'pt', format: 'letter', orientation: 'portrait' })

    for (let p = 0; p < pages.length; p++) {
      if (p > 0) pdf.addPage()
      pdf.setFillColor(17, 17, 20)
      pdf.rect(0, 0, pdfPageWidth, pdfPageHeight, 'F')

      const { startPx, endPx } = pages[p]
      const sliceH = endPx - startPx
      if (sliceH <= 0) continue

      const srcY = Math.min(Math.round(startPx * canvasScale), canvas.height - 1)
      const srcH = Math.min(Math.round(sliceH * canvasScale), canvas.height - srcY)
      if (srcH <= 0) continue

      const destH = sliceH * cssToPdf
      const pageCanvas = document.createElement('canvas')
      pageCanvas.width = canvas.width
      pageCanvas.height = srcH
      const imageData = canvas.getContext('2d').getImageData(0, srcY, canvas.width, srcH)
      pageCanvas.getContext('2d').putImageData(imageData, 0, 0)

      pdf.addImage(pageCanvas.toDataURL('image/jpeg', 0.95), 'JPEG', margin, margin, contentWidth, destH)
    }

    pdf.save(`${baseName.value}.pdf`)
    toast('PDF exported')
  } catch (e) {
    console.error('PDF export failed:', e)
    toast('PDF export failed')
  }
}

async function exportPng() {
  closeExport()
  toast('Generating PNG...')
  try {
    const { canvas } = await renderToCanvas()
    canvas.toBlob((blob) => {
      download(blob, `${baseName.value}.png`, 'image/png')
      toast('PNG exported')
    }, 'image/png')
  } catch (e) {
    console.error('PNG export failed:', e)
    toast('PNG export failed')
  }
}

function exportHtml() {
  closeExport()
  const out = isFullDoc.value ? source.value : frameDoc.value
  download(out, `${baseName.value}.html`, 'text/html')
  toast('HTML exported')
}

const defaultContent = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>HTML Renderer</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 720px; margin: 2rem auto; padding: 0 1rem; background: #111114; color: #e4e4e8; line-height: 1.7; }
  h1 { color: #c4956a; border-bottom: 1px solid #2a2a30; padding-bottom: 0.5rem; }
  code { background: #0a0a0c; padding: 2px 6px; border-radius: 4px; font-family: ui-monospace, monospace; color: #c4956a; }
  pre { background: #0a0a0c; padding: 1rem; border-radius: 6px; overflow: auto; }
  blockquote { border-left: 3px solid #c4956a; margin: 1rem 0; padding: 0.5rem 1rem; color: #8e8e96; font-style: italic; }
  button { background: #c4956a; color: #111114; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-weight: 600; }
  button:hover { background: #a07850; }
</style>
</head>
<body>
  <h1>HTML Renderer</h1>
  <p>Live preview of any HTML you paste or type. Full <code>&lt;script&gt;</code>, <code>&lt;style&gt;</code>, and <code>&lt;link&gt;</code> support inside a sandboxed iframe.</p>

  <h2>Try it</h2>
  <button onclick="alert('Scripts run in the sandbox!')">Click me</button>

  <h2>Code</h2>
  <pre><code>const tool = document.querySelector('.tool');
tool.render();</code></pre>

  <blockquote>Drop a <code>.html</code> file, paste markup, or just type. Export to PDF, PNG, or HTML.</blockquote>
</body>
</html>
`

const { useShareable } = useShare()
let defaultLoaded = false

useShareable('html', (shared) => {
  source.value = shared.source || ''
  defaultLoaded = true
})

onMounted(() => {
  if (!defaultLoaded) {
    source.value = defaultContent
  }
})
</script>

<style scoped>
.toolbar-filename {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  color: var(--text-muted);
  padding: 4px 10px;
  background: var(--bg-deep);
  border-radius: var(--radius);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  opacity: 0;
  transition: opacity var(--transition);
}
.toolbar-filename.visible { opacity: 1; }

.preview-full {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preview-scroll {
  flex: 1;
  overflow: hidden;
  display: flex;
}

.html-frame {
  flex: 1;
  width: 100%;
  height: 100%;
  border: none;
  background: #111114;
}

.export-wrapper { position: relative; }

.export-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 180px;
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

.export-menu.open {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.export-item {
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

.export-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.export-icon {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  color: var(--text-muted);
  width: 30px;
  text-align: right;
  flex-shrink: 0;
  letter-spacing: -0.02em;
}

.export-sep {
  height: 1px;
  background: var(--border-subtle);
  margin: 4px 0;
}
</style>
