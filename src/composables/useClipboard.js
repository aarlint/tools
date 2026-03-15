import { useToast } from './useToast'

export function useClipboard() {
  const { toast } = useToast()

  async function copy(text, label = 'Copied') {
    try {
      await navigator.clipboard.writeText(text)
      toast(label)
    } catch {
      toast('Copy failed')
    }
  }

  return { copy }
}
