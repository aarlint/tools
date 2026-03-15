<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <span class="toolbar-filename" :class="{ visible: filename }">{{ filename }}</span>
      <span class="tag">{{ wordCount }} words</span>
      <span class="tag">{{ charCount }} chars</span>
      <div class="spacer"></div>
      <button class="toolbar-btn" :class="{ active: viewMode === 'split' }" @click="viewMode = 'split'">Split</button>
      <button class="toolbar-btn" :class="{ active: viewMode === 'preview' }" @click="viewMode = 'preview'">Preview</button>
      <ShareButton tool="md" :getState="() => ({ source: source })" />
      <div class="export-wrapper">
        <button class="toolbar-btn" :class="{ active: exportOpen }" @click.stop="exportOpen = !exportOpen">Export</button>
        <div class="export-menu" :class="{ open: exportOpen }" @click.stop>
          <button class="export-item" @click="exportPdf"><span class="export-icon">.pdf</span> PDF Document</button>
          <button class="export-item" @click="exportPng"><span class="export-icon">.png</span> PNG Image</button>
          <div class="export-sep"></div>
          <button class="export-item" @click="exportHtml"><span class="export-icon">.html</span> HTML File</button>
          <button class="export-item" @click="download(source, baseName + '.md', 'text/markdown'); closeExport(); toast('Markdown exported')"><span class="export-icon">.md</span> Markdown</button>
          <div class="export-sep"></div>
          <button class="export-item" @click="copy(mdHtml, 'HTML copied'); closeExport()"><span class="export-icon">clipboard</span> Copy HTML</button>
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
              placeholder="Write or paste markdown here..."
              spellcheck="false"
              @keydown.tab.prevent="insertTab"
            ></textarea>
          </div>
        </template>
        <template #right>
          <div class="pane preview-pane">
            <PaneHeader label="Preview">{{ wordCount }} words</PaneHeader>
            <div class="preview-scroll">
              <div class="md-body" ref="mdBodyEl" v-html="mdHtml"></div>
            </div>
          </div>
        </template>
      </SplitPane>
      <div v-else class="pane preview-pane preview-full">
        <PaneHeader label="Preview">{{ wordCount }} words</PaneHeader>
        <div class="preview-scroll">
          <div class="md-body" ref="mdBodyEl" v-html="mdHtml"></div>
        </div>
      </div>
    </div>
    <DropOverlay :active="dragging" text="Drop your .md file" hint=".md · .markdown · .txt" />
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { marked, Renderer } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark-dimmed.css'
import mermaid from 'mermaid'
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
const mdHtml = ref('')
const editorEl = ref(null)
const mdBodyEl = ref(null)

const { dragging } = useFileDrop((content, name) => {
  source.value = content
  filename.value = name
  toast(`Loaded ${name}`)
}, { extensions: ['.md', '.markdown', '.txt'] })

const charCount = computed(() => source.value.length.toLocaleString())
const wordCount = computed(() => {
  const t = source.value.trim()
  return t ? t.split(/\s+/).length.toLocaleString() : '0'
})

// Mermaid setup
mermaid.initialize({
  startOnLoad: false,
  theme: 'dark',
  themeVariables: {
    darkMode: true,
    background: '#0a0a0c',
    primaryColor: '#2a2a35',
    primaryTextColor: '#e4e4e8',
    primaryBorderColor: '#3a3a45',
    lineColor: '#5a5a64',
    secondaryColor: '#1a1a22',
    tertiaryColor: '#14141a',
    noteBkgColor: '#1e1e28',
    noteTextColor: '#c4956a',
    fontFamily: 'DM Sans, sans-serif',
  },
  flowchart: { curve: 'basis', padding: 20 },
  sequence: { actorMargin: 50, messageMargin: 40 },
  fontSize: 14,
})

// Marked setup
let mermaidId = 0
const renderer = new Renderer()

renderer.code = function({ text, lang }) {
  if (lang === 'mermaid') {
    const id = `mermaid-${mermaidId++}`
    return `<div class="mermaid-container" data-mermaid-id="${id}" data-mermaid-src="${encodeURIComponent(text)}"></div>`
  }
  const validLang = lang && hljs.getLanguage(lang) ? lang : null
  const highlighted = validLang
    ? hljs.highlight(text, { language: lang }).value
    : hljs.highlightAuto(text).value
  const langLabel = lang ? `<span class="code-lang">${lang}</span>` : ''
  return `<pre>${langLabel}<code class="hljs${validLang ? ` language-${lang}` : ''}">${highlighted}</code></pre>`
}

renderer.listitem = function({ text, task, checked }) {
  if (task) {
    return `<li class="task-list-item"><input type="checkbox" disabled ${checked ? 'checked' : ''}><span>${text}</span></li>`
  }
  return `<li>${text}</li>`
}

marked.setOptions({ renderer, gfm: true, breaks: false })

// Render
let renderTimer = null
async function render() {
  mermaidId = 0
  mdHtml.value = marked.parse(source.value)

  await nextTick()

  // Render mermaid diagrams
  if (!mdBodyEl.value) return
  const els = mdBodyEl.value.querySelectorAll('.mermaid-container[data-mermaid-src]')
  for (const el of els) {
    const src = decodeURIComponent(el.dataset.mermaidSrc)
    const id = el.dataset.mermaidId
    try {
      const { svg } = await mermaid.render(id, src)
      el.innerHTML = svg
    } catch (e) {
      el.innerHTML = `<div class="mermaid-error">Mermaid error: ${e.message || 'Invalid syntax'}</div>`
      const errEl = document.getElementById('d' + id)
      if (errEl) errEl.remove()
    }
  }
}

watch(source, () => {
  clearTimeout(renderTimer)
  renderTimer = setTimeout(render, 150)
})

watch(viewMode, () => {
  nextTick(render)
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

// Export
const baseName = computed(() => filename.value?.replace(/\.[^.]+$/, '') || 'document')

function closeExport() { exportOpen.value = false }

// Close export menu on outside click
onMounted(() => {
  document.addEventListener('click', () => { exportOpen.value = false })
})

async function renderToCanvas() {
  const container = document.createElement('div')
  container.style.cssText = `
    position: absolute; left: -9999px; top: 0;
    width: 816px;
    background: #111114;
    padding: 48px 56px;
    font-family: ${getComputedStyle(mdBodyEl.value).fontFamily};
  `
  container.className = 'md-body'
  container.innerHTML = mdBodyEl.value.innerHTML
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
  const styles = document.querySelector('style')?.textContent || ''
  const html = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>${baseName.value}</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@300;400;500;600&family=Instrument+Serif:ital@0;1&family=DM+Sans:opsz,wght@9..40,300;9..40,400;9..40,500;9..40,600&display=swap" rel="stylesheet">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark-dimmed.min.css">
<style>
  :root { --bg-deep:#0a0a0c;--bg-surface:#111114;--bg-raised:#18181c;--border:#2a2a30;--border-subtle:#1e1e24;--text-primary:#e4e4e8;--text-secondary:#8e8e96;--text-muted:#5a5a64;--accent:#c4956a;--accent-dim:#a07850;--accent-glow:rgba(196,149,106,0.08);--green:#7ec89b;--red:#d47979;--blue:#7ca5d4;--purple:#b49cdb;--font-mono:'IBM Plex Mono',monospace;--font-serif:'Instrument Serif',Georgia,serif;--font-sans:'DM Sans',sans-serif;--radius:6px; }
  body { background:var(--bg-surface); color:var(--text-primary); font-family:var(--font-sans); margin:0; padding:48px 56px; display:flex; justify-content:center; }
  .md-body { max-width:860px; width:100%; font-size:0.95rem; line-height:1.75; word-wrap:break-word; }
  ${styles}
</style>
</head>
<body>
<div class="md-body">${mdBodyEl.value.innerHTML}</div>
</body>
</html>`
  download(html, `${baseName.value}.html`, 'text/html')
  toast('HTML exported')
}

// Default content
const defaultContent = `# Welcome to tools.

A dark, refined **Markdown** & **Mermaid** previewer. Drag a file, paste text, or just start typing.

---

## Features

- [x] Full **GFM** Markdown support
- [x] Mermaid diagram rendering
- [x] Syntax-highlighted code blocks
- [x] Drag & drop file loading
- [ ] Your next great document

## Code

\`\`\`javascript
async function render(markdown) {
  const html = marked.parse(markdown);
  document.body.innerHTML = html;
  await mermaid.run();
}
\`\`\`

## Diagrams

\`\`\`mermaid
graph LR
    A[Markdown] --> B{Parser}
    B --> C[HTML]
    B --> D[Mermaid]
    D --> E[SVG Diagram]
    C --> F[Rendered Preview]
    E --> F
    style A fill:#2a2a35,stroke:#c4956a,color:#e4e4e8
    style F fill:#2a2a35,stroke:#7ec89b,color:#e4e4e8
\`\`\`

## Tables

| Feature | Status | Notes |
|---------|--------|-------|
| GFM Markdown | Active | Full spec support |
| Mermaid | Active | Flowcharts, sequences, etc. |
| Syntax Highlighting | Active | 190+ languages |

> *"The art of programming is the art of organizing complexity."*
> — Edsger W. Dijkstra

Inline \`code\`, **bold**, *italic*, and ~~strikethrough~~.
`

const { useShareable } = useShare()
let defaultLoaded = false

useShareable('md', (shared) => {
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
  overflow-y: auto;
  padding: 32px 40px;
}

.export-wrapper {
  position: relative;
}

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
