import { useState, useEffect } from 'react'
import Popup from './components/Popup'
import Settings from './components/Settings'
import Welcome from './components/Welcome'
import MiniMode from './components/MiniMode'
import DiffView from './components/DiffView'
import * as runtime from '../wailsjs/runtime'
import * as AppAPI from '../wailsjs/go/main/App'
import { config as configModels, rewriter as rewriterModels } from '../wailsjs/go/models'
import './styles/main.css'

export type View = 'popup' | 'settings' | 'welcome' | 'mini' | 'diff'

export type Config = configModels.Config

export interface TextTypeInfo {
  Type: string
  Label: string
  Icon: string
  Description: string
}

export interface StyleInfo {
  Label: string
  Icon: string
  Description: string
}

function App() {
  const [currentView, setCurrentView] = useState<View>('settings')
  const [selectedText, setSelectedText] = useState('')
  const [selectionTrigger, setSelectionTrigger] = useState(0)
  const [settings, setSettings] = useState<Config | null>(null)
  const [miniModeResult, setMiniModeResult] = useState<string>('')
  const [isGenerating, setIsGenerating] = useState(false)
  const [diffOriginal, setDiffOriginal] = useState('')
  const [diffRewritten, setDiffRewritten] = useState('')

  useEffect(() => {
    // Listen for text selection event from backend
    runtime.EventsOn('text:selected', async (text: string) => {
      console.log('Frontend received text:', text)
      setSelectedText(text)
      setSelectionTrigger(prev => prev + 1)
      setMiniModeResult('')
      
      try {
        const config = await AppAPI.GetSettings()
        const useMini = config?.mini_mode ?? false
        setCurrentView(prev => {
          if (prev === 'welcome') return 'welcome'
          return useMini ? 'mini' : 'popup'
        })
      } catch (e) {
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
      const config = await AppAPI.GetSettings()
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
      await AppAPI.ApplyRewrite(text)
      runtime.WindowHide()
    } catch (error) {
      console.error('Failed to apply rewrite:', error)
    }
  }

const handleSaveSettings = async (newSettings: Config) => {
    try {
      await AppAPI.SaveSettings(newSettings)
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
      const prompts = await AppAPI.GetAllCustomPrompts()
      return prompts || {}
    } catch (error) {
      console.error('Failed to load custom prompts:', error)
      return {}
    }
  }

  const saveCustomPrompt = async (style: string, textType: string, prompt: string): Promise<{ success: boolean; error?: string }> => {
    try {
      await AppAPI.SetCustomPrompt(style, textType, prompt)
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to save custom prompt'
      console.error('Failed to save custom prompt:', error)
      return { success: false, error: errorMsg }
    }
  }

  const deleteCustomPrompt = async (style: string, textType: string): Promise<{ success: boolean; error?: string }> => {
    try {
      await AppAPI.DeleteCustomPrompt(style, textType)
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to delete custom prompt'
      console.error('Failed to delete custom prompt:', error)
      return { success: false, error: errorMsg }
    }
  }

  const resetAllCustomPrompts = async (): Promise<{ success: boolean; error?: string }> => {
    try {
      await AppAPI.ResetAllCustomPrompts()
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to reset all prompts'
      console.error('Failed to reset all prompts:', error)
      return { success: false, error: errorMsg }
    }
  }

  const getDefaultPrompt = async (style: string, textType: string): Promise<string> => {
    try {
      const prompt = await AppAPI.GetDefaultPrompt(style, textType)
      return prompt
    } catch (error) {
      console.error('Failed to get default prompt:', error)
      return ''
    }
  }

  const loadRewriteStyles = async (): Promise<string[]> => {
    try {
      const styles = await AppAPI.GetRewriteStyles()
      return styles || []
    } catch (error) {
      console.error('Failed to load rewrite styles:', error)
      return []
    }
  }

  const loadAnalysisStyles = async (): Promise<string[]> => {
    try {
      const styles = await AppAPI.GetAnalysisStyles()
      return styles || []
    } catch (error) {
      console.error('Failed to load analysis styles:', error)
      return []
    }
  }

  const loadTextTypes = async (): Promise<rewriterModels.TextTypeInfo[]> => {
    try {
      const types = await AppAPI.GetTextTypes()
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
      const result = await AppAPI.RetryRewriteWithFormatting(selectedText, style, true)
      if (result.text) {
        setMiniModeResult(result.text)
        await AppAPI.ApplyRewrite(result.text)
      }
    } catch (error) {
      console.error('Mini mode rewrite failed:', error)
    }
    setIsGenerating(false)
  }

  const handleShowDiff = (original: string, rewritten: string) => {
    setDiffOriginal(original)
    setDiffRewritten(rewritten)
    setCurrentView('diff')
  }

  const [lastResult, setLastResult] = useState('')

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
          onShowDiff={(rewritten) => handleShowDiff(selectedText, rewritten)}
          defaultStyle={settings.default_style}
          miniModeResult={miniModeResult}
          onResultChange={setLastResult}
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
          onResetAllCustomPrompts={resetAllCustomPrompts}
          onGetDefaultPrompt={getDefaultPrompt}
          onLoadRewriteStyles={loadRewriteStyles}
          onLoadAnalysisStyles={loadAnalysisStyles}
          onLoadTextTypes={loadTextTypes}
        />
      )}

      {currentView === 'diff' && (
        <DiffView
          originalText={diffOriginal}
          rewrittenText={diffRewritten}
          onClose={() => setCurrentView(selectedText ? 'popup' : 'settings')}
        />
      )}
    </div>
  )
}

export default App
