import { ref, onMounted, onUnmounted } from 'vue'

export function useFileDrop(onFile, { extensions = [] } = {}) {
  const dragging = ref(false)
  let dragCounter = 0

  function onDragEnter(e) {
    e.preventDefault()
    dragCounter++
    if (dragCounter === 1) dragging.value = true
  }

  function onDragLeave(e) {
    e.preventDefault()
    dragCounter--
    if (dragCounter <= 0) {
      dragCounter = 0
      dragging.value = false
    }
  }

  function onDragOver(e) {
    e.preventDefault()
  }

  function onDrop(e) {
    e.preventDefault()
    dragCounter = 0
    dragging.value = false

    const file = e.dataTransfer?.files?.[0]
    if (!file) return

    if (extensions.length) {
      const ext = '.' + file.name.split('.').pop().toLowerCase()
      if (!extensions.includes(ext)) return
    }

    const reader = new FileReader()
    reader.onload = (ev) => {
      onFile(ev.target.result, file.name)
    }
    reader.readAsText(file)
  }

  onMounted(() => {
    document.addEventListener('dragenter', onDragEnter)
    document.addEventListener('dragleave', onDragLeave)
    document.addEventListener('dragover', onDragOver)
    document.addEventListener('drop', onDrop)
  })

  onUnmounted(() => {
    document.removeEventListener('dragenter', onDragEnter)
    document.removeEventListener('dragleave', onDragLeave)
    document.removeEventListener('dragover', onDragOver)
    document.removeEventListener('drop', onDrop)
  })

  return { dragging }
}
