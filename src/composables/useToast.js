import { ref } from 'vue'

const message = ref('')
const visible = ref(false)
let timeout = null

export function useToast() {
  function toast(msg, duration = 2200) {
    message.value = msg
    visible.value = true
    clearTimeout(timeout)
    timeout = setTimeout(() => { visible.value = false }, duration)
  }

  return { message, visible, toast }
}
