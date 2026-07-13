import { useState, useCallback } from 'react'
import * as runtime from '../../wailsjs/runtime'
import * as AppAPI from '../../wailsjs/go/main/App'
import * as SettingsAPI from '../../wailsjs/go/main/SettingsService'

interface UseClipboardPasteProps {
  result: string
  onClose: () => void
  autoPasteMode: string
  setAutoPasteMode: (mode: string) => void
}

export function useClipboardPaste({
  result,
  onClose,
  autoPasteMode,
  setAutoPasteMode
}: UseClipboardPasteProps) {
  const [copied, setCopied] = useState(false)
  const [showPasteDialog, setShowPasteDialog] = useState(false)
  const [dontAskAgain, setDontAskAgain] = useState(false)

  const handleCopy = useCallback(async () => {
    if (!result) return

    try {
      await AppAPI.ApplyRewrite(result)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)

      const settings = await SettingsAPI.GetSettings()
      if (settings?.auto_minimize_on_copy) {
        setTimeout(() => {
          runtime.WindowMinimise()
        }, 400)
      }
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }, [result])

  const handleReplace = useCallback(async () => {
    if (!result) return

    if (autoPasteMode === 'ask') {
      setShowPasteDialog(true)
    } else if (autoPasteMode === 'always') {
      try {
        await AppAPI.ApplyRewriteAndPaste(result)
        onClose()
      } catch (err) {
        console.error('Failed to paste:', err)
        await handleCopy()
        onClose()
      }
    } else {
      await handleCopy()
      onClose()
    }
  }, [result, autoPasteMode, onClose, handleCopy])

  const handlePasteConfirm = useCallback(async () => {
    if (dontAskAgain) {
      try {
        const settings = await SettingsAPI.GetSettings()
        settings.auto_paste_mode = 'always'
        await SettingsAPI.SaveSettings(settings)
        setAutoPasteMode('always')
      } catch (e) {
        console.error('Failed to save settings:', e)
      }
    }

    setShowPasteDialog(false)
    try {
      await AppAPI.ApplyRewriteAndPaste(result)
      onClose()
    } catch (err) {
      console.error('Failed to paste:', err)
      await handleCopy()
      onClose()
    }
  }, [dontAskAgain, result, onClose, handleCopy, setAutoPasteMode])

  const handlePasteCancel = useCallback(async () => {
    if (dontAskAgain) {
      try {
        const settings = await SettingsAPI.GetSettings()
        settings.auto_paste_mode = 'never'
        await SettingsAPI.SaveSettings(settings)
        setAutoPasteMode('never')
      } catch (e) {
        console.error('Failed to save settings:', e)
      }
    }

    setShowPasteDialog(false)
    await handleCopy()
    onClose()
  }, [dontAskAgain, handleCopy, onClose, setAutoPasteMode])

  return {
    copied,
    showPasteDialog,
    setShowPasteDialog,
    dontAskAgain,
    setDontAskAgain,
    handleCopy,
    handleReplace,
    handlePasteConfirm,
    handlePasteCancel
  }
}
