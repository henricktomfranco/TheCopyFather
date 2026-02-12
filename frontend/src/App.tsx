import { useState, useEffect } from 'react'
import Popup from './components/Popup'
import Settings from './components/Settings'
import Welcome from './components/Welcome'
import MiniMode from './components/MiniMode'
import * as runtime from '../wailsjs/runtime'
import './styles/main.css'

export type View = 'popup' | 'settings' | 'welcome' | 'mini'

export interface RewriteOption {
  style: string
  text: string
  error?: string
}

export interface Config {
  server_url: string
  model: string
  api_key?: string
  default_style: string
  auto_start: boolean
  hotkey: string
  monitor_clipboard: boolean
  first_run: boolean
  auto_paste_mode?: string
  popup_position_mode?: string
  custom_prompts?: Record<string, Record<string, string>>
  mini_mode?: boolean
}

export interface TextTypeInfo {
  type: string
  label: string
  icon: string
  description: string
}

export interface StyleInfo {
  label: string
  icon: string
  description: string
}

function App() {
  const [currentView, setCurrentView] = useState<View>('settings')
  const [selectedText, setSelectedText] = useState('')
  const [selectionTrigger, setSelectionTrigger] = useState(0)
  const [settings, setSettings] = useState<Config | null>(null)
  const [miniModeResult, setMiniModeResult] = useState<string>('')
  const [isGenerating, setIsGenerating] = useState(false)

  useEffect(() => {
    // Listen for text selection event from backend
    runtime.EventsOn('text:selected', async (text: string) => {
      console.log('Frontend received text:', text)
      setSelectedText(text)
      setSelectionTrigger(prev => prev + 1)
      setMiniModeResult('')
      
      // Read mini_mode setting directly from backend
      try {
        // @ts-ignore
        const config = await window.go.main.App.GetSettings()
        const useMini = config?.mini_mode ?? false
        setCurrentView(prev => {
          if (prev === 'welcome') return 'welcome'
          return useMini ? 'mini' : 'popup'
        })
      } catch (e) {
        // Fallback to popup if settings can't be read
        setCurrentView('popup')
      }
    })

    // Listen for show settings event from system tray
    runtime.EventsOn('window:showsettings', () => {
      setCurrentView('settings')
      loadSettings()
    })

    // Load initial settings
    loadSettings()

    return () => {
      runtime.EventsOff('text:selected')
      runtime.EventsOff('window:showsettings')
    }
  }, [])

  const loadSettings = async () => {
    try {
      // @ts-ignore
      const config = await window.go.main.App.GetSettings()
      setSettings(config)
      if (config.first_run) {
        setCurrentView('welcome')
      }
    } catch (error) {
      console.error('Failed to load settings:', error)
    }
  }

  const handleSelectRewrite = async (text: string) => {
    try {
      // @ts-ignore
      await window.go.main.App.ApplyRewrite(text)
      runtime.WindowHide()
    } catch (error) {
      console.error('Failed to apply rewrite:', error)
    }
  }

const handleSaveSettings = async (newSettings: Config) => {
    try {
      // @ts-ignore
      await window.go.main.App.SaveSettings(newSettings)
      setSettings(newSettings)
      if (selectedText) {
        setCurrentView('popup')
      }
    } catch (error) {
      console.error('Failed to save settings:', error)
    }
  }

  const loadCustomPrompts = async (): Promise<Record<string, Record<string, string>>> => {
    try {
      // @ts-ignore
      const prompts = await window.go.main.App.GetAllCustomPrompts()
      return prompts || {}
    } catch (error) {
      console.error('Failed to load custom prompts:', error)
      return {}
    }
  }

  const saveCustomPrompt = async (style: string, textType: string, prompt: string): Promise<{ success: boolean; error?: string }> => {
    try {
      // @ts-ignore
      await window.go.main.App.SetCustomPrompt(style, textType, prompt)
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to save custom prompt'
      console.error('Failed to save custom prompt:', error)
      return { success: false, error: errorMsg }
    }
  }

  const deleteCustomPrompt = async (style: string, textType: string): Promise<{ success: boolean; error?: string }> => {
    try {
      // @ts-ignore
      await window.go.main.App.DeleteCustomPrompt(style, textType)
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to delete custom prompt'
      console.error('Failed to delete custom prompt:', error)
      return { success: false, error: errorMsg }
    }
  }

  const getDefaultPrompt = async (style: string, textType: string): Promise<string> => {
    try {
      // @ts-ignore
      const prompt = await window.go.main.App.GetDefaultPrompt(style, textType)
      return prompt
    } catch (error) {
      console.error('Failed to get default prompt:', error)
      return ''
    }
  }

  const loadRewriteStyles = async (): Promise<string[]> => {
    try {
      // @ts-ignore
      const styles = await window.go.main.App.GetRewriteStyles()
      return styles || []
    } catch (error) {
      console.error('Failed to load rewrite styles:', error)
      return []
    }
  }

  const loadAnalysisStyles = async (): Promise<string[]> => {
    try {
      // @ts-ignore
      const styles = await window.go.main.App.GetAnalysisStyles()
      return styles || []
    } catch (error) {
      console.error('Failed to load analysis styles:', error)
      return []
    }
  }

  const loadTextTypes = async (): Promise<TextTypeInfo[]> => {
    try {
      // @ts-ignore
      const types = await window.go.main.App.GetTextTypes()
      return types || []
    } catch (error) {
      console.error('Failed to load text types:', error)
      return []
    }
  }

  const handleAcceptFirstRun = async () => {
    if (!settings) return
    const newSettings = { ...settings, first_run: false }
    await handleSaveSettings(newSettings)
    setCurrentView('settings') // Go to settings first to confirm URL
  }

  const handleClose = () => {
    runtime.WindowHide()
  }

  const handleOpenSettings = () => {
    setCurrentView('settings')
  }

  const handleMiniModeExpand = () => {
    setCurrentView('popup')
  }

  const handleMiniModeRewrite = async (style: string) => {
    if (!selectedText) return
    setIsGenerating(true)
    try {
      // @ts-ignore
      const result = await window.go.main.App.RetryRewriteWithFormatting(selectedText, style, true)
      if (result.text) {
        setMiniModeResult(result.text)
        // Auto-copy to clipboard
        // @ts-ignore
        await window.go.main.App.ApplyRewrite(result.text)
      }
    } catch (error) {
      console.error('Mini mode rewrite failed:', error)
    }
    setIsGenerating(false)
  }

  const rewriteStylesList = [
    { value: 'grammar', label: 'Grammar', icon: '🛡️' },
    { value: 'standard', label: 'Standard', icon: '📝' },
    { value: 'formal', label: 'Formal', icon: '📢' },
    { value: 'casual', label: 'Casual', icon: '💬' },
    { value: 'creative', label: 'Creative', icon: '✨' },
    { value: 'short', label: 'Short', icon: '📏' },
    { value: 'expand', label: 'Expand', icon: '📖' },
  ]

  return (
    <div className="app">
      {currentView === 'welcome' && (
        <Welcome onAccept={handleAcceptFirstRun} />
      )}

      {currentView === 'mini' && settings && (
        <MiniMode
          originalText={selectedText}
          currentStyle={settings.default_style}
          onExpand={handleMiniModeExpand}
          onClose={handleClose}
          onStyleChange={(style) => {
            const newSettings = { ...settings, default_style: style }
            handleSaveSettings(newSettings)
          }}
          onQuickRewrite={() => handleMiniModeRewrite(settings.default_style)}
          availableStyles={rewriteStylesList}
          isGenerating={isGenerating}
        />
      )}

      {currentView === 'popup' && settings && (
        <Popup
          key={`popup-${selectionTrigger}`}
          originalText={selectedText}
          onSelect={handleSelectRewrite}
          onClose={handleClose}
          onSettings={handleOpenSettings}
          defaultStyle={settings.default_style}
          miniModeResult={miniModeResult}
        />
      )}

      {currentView === 'settings' && settings && (
        <Settings
          settings={settings}
          onSave={handleSaveSettings}
          onCancel={() => {
            if (selectedText) {
              setCurrentView('popup')
            } else {
              handleClose()
            }
          }}
          onLoadCustomPrompts={loadCustomPrompts}
          onSaveCustomPrompt={saveCustomPrompt}
          onDeleteCustomPrompt={deleteCustomPrompt}
          onGetDefaultPrompt={getDefaultPrompt}
          onLoadRewriteStyles={loadRewriteStyles}
          onLoadAnalysisStyles={loadAnalysisStyles}
          onLoadTextTypes={loadTextTypes}
        />
      )}
    </div>
  )
}

export default App
