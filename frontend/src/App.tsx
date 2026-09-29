import { useState, useEffect } from 'react'
import Popup from './components/Popup'
import Settings from './components/Settings'
import Welcome from './components/Welcome'
import MiniMode from './components/MiniMode'
import DiffView from './components/DiffView'
import * as runtime from '../wailsjs/runtime'
import * as AppAPI from '../wailsjs/go/main/App'
import * as SettingsAPI from '../wailsjs/go/main/SettingsService'
import * as RewriteAPI from '../wailsjs/go/main/RewriteService'
import { config as configModels, rewriter as rewriterModels } from '../wailsjs/go/models'
import './styles/main.css'

export type View = 'popup' | 'settings' | 'welcome' | 'mini' | 'diff'

// Config type for app state (all fields optional to allow partial updates)
export interface Config {
  // Provider mode: "embedded" | "openai"
  provider_mode?: string;
  // Embedded llama.cpp settings
  embedded_model?: string;
  embedded_model_path?: string;
  embedded_binary_path?: string;
  embedded_hardware?: string;
  embedded_cpu_threads?: number;
  embedded_max_threads?: number;
  embedded_context_size?: number;
  disable_thinking?: boolean;
  // From configModels.Config
  server_url?: string;
  model?: string;
  api_key?: string;
  default_style?: string;
  auto_start?: boolean;
  hotkey?: string;
  monitor_clipboard?: boolean;
  first_run?: boolean;
  custom_prompts?: Record<string, any>;
  auto_paste_mode?: string;
  popup_position_mode?: string;
  mini_mode?: boolean;
  auto_minimize_on_copy?: boolean;
  // Ghost mode settings
  ghost_hotkey?: string;
  ghost_style?: string;
  ghost_text_type?: string;
  // Auto-update settings
  autoUpdateEnabled?: boolean;
  currentVersion?: string;
  updateChannel?: string;
  // OpenAI-compatible settings
  useOpenAICompatible?: boolean;
  openAIBaseURL?: string;
  openAIModel?: string;
  openAIAPIKey?: string;
  // Streaming settings
  disableStreaming?: boolean;
}

// Helper to convert partial Config to full Config
const toFullConfig = (config: Config): configModels.Config => {
  const mode = config.provider_mode || (config.useOpenAICompatible ? 'openai' : 'embedded');
  const isOpenAI = mode === 'openai' || !!config.useOpenAICompatible;
  return {
    provider_mode: mode,
    embedded_model: config.embedded_model ?? 'SmolLM2-360M',
    embedded_model_path: config.embedded_model_path ?? 'models/smollm2-360m-instruct-q4_k_m.gguf',
    embedded_binary_path: config.embedded_binary_path ?? 'engine/llama-server.exe',
    embedded_hardware: config.embedded_hardware ?? 'cpu',
    embedded_cpu_threads: config.embedded_cpu_threads ?? 0,
    embedded_max_threads: config.embedded_max_threads ?? 6,
    embedded_context_size: config.embedded_context_size ?? 2048,
    disable_thinking: config.disable_thinking ?? true,
    autoUpdateEnabled: config.autoUpdateEnabled ?? false,
    currentVersion: config.currentVersion ?? '',
    updateChannel: config.updateChannel ?? '',
    useOpenAICompatible: isOpenAI,
    openAIBaseURL: config.openAIBaseURL ?? 'https://integrate.api.nvidia.com/v1',
    openAIModel: config.openAIModel ?? 'mistralai/mistral-7b-instruct',
    openAIAPIKey: config.openAIAPIKey ?? '',
    server_url: config.server_url ?? '',
    model: config.model ?? 'SmolLM2-360M',
    api_key: config.api_key,
    disableStreaming: config.disableStreaming ?? false,
    default_style: config.default_style ?? '',
    auto_start: config.auto_start ?? false,
    hotkey: config.hotkey ?? '',
    monitor_clipboard: config.monitor_clipboard ?? false,
    first_run: config.first_run ?? false,
    auto_paste_mode: config.auto_paste_mode ?? '',
    popup_position_mode: config.popup_position_mode ?? '',
    mini_mode: config.mini_mode ?? false,
    auto_minimize_on_copy: config.auto_minimize_on_copy ?? false,
    ghost_hotkey: config.ghost_hotkey ?? '',
    ghost_style: config.ghost_style ?? '',
    ghost_text_type: config.ghost_text_type ?? '',
  };
};

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
        const config = await SettingsAPI.GetSettings()
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
      const config = await SettingsAPI.GetSettings()
      setSettings(config)

      // On launch, if running embedded mode and missing files, show setup screen
      try {
        const setup = await SettingsAPI.CheckSetupStatus()
        const mode = config?.provider_mode || (config?.useOpenAICompatible ? 'openai' : 'embedded')
        if (mode === 'embedded' && (!setup.engine_installed || !setup.model_installed)) {
          setCurrentView('welcome')
          return
        }
      } catch (e) {
        console.error('Failed to check setup status:', e)
      }

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
await SettingsAPI.SaveSettings(toFullConfig(newSettings))
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
      const prompts = await RewriteAPI.GetAllCustomPrompts()
      return prompts || {}
    } catch (error) {
      console.error('Failed to load custom prompts:', error)
      return {}
    }
  }

  const saveCustomPrompt = async (style: string, textType: string, prompt: string): Promise<{ success: boolean; error?: string }> => {
    try {
      await RewriteAPI.SetCustomPrompt(style, textType, prompt)
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to save custom prompt'
      console.error('Failed to save custom prompt:', error)
      return { success: false, error: errorMsg }
    }
  }

  const deleteCustomPrompt = async (style: string, textType: string): Promise<{ success: boolean; error?: string }> => {
    try {
      await RewriteAPI.DeleteCustomPrompt(style, textType)
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to delete custom prompt'
      console.error('Failed to delete custom prompt:', error)
      return { success: false, error: errorMsg }
    }
  }

  const resetAllCustomPrompts = async (): Promise<{ success: boolean; error?: string }> => {
    try {
      await RewriteAPI.ResetAllCustomPrompts()
      return { success: true }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to reset all prompts'
      console.error('Failed to reset all prompts:', error)
      return { success: false, error: errorMsg }
    }
  }

  const getDefaultPrompt = async (style: string, textType: string): Promise<string> => {
    try {
      const prompt = await RewriteAPI.GetDefaultPrompt(style, textType)
      return prompt
    } catch (error) {
      console.error('Failed to get default prompt:', error)
      return ''
    }
  }

  const loadRewriteStyles = async (): Promise<string[]> => {
    try {
      const styles = await RewriteAPI.GetRewriteStyles()
      return styles || []
    } catch (error) {
      console.error('Failed to load rewrite styles:', error)
      return []
    }
  }

  const loadAnalysisStyles = async (): Promise<string[]> => {
    try {
      const styles = await RewriteAPI.GetAnalysisStyles()
      return styles || []
    } catch (error) {
      console.error('Failed to load analysis styles:', error)
      return []
    }
  }

  const loadTextTypes = async (): Promise<rewriterModels.TextTypeInfo[]> => {
    try {
      const types = await RewriteAPI.GetTextTypes()
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
    const requestID = crypto.randomUUID()
    let resultText = ''

    const cleanup = () => {
      runtime.EventsOff(`stream:chunk:${requestID}`)
      runtime.EventsOff(`stream:done:${requestID}`)
      runtime.EventsOff(`stream:error:${requestID}`)
    }

    try {
      await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => {
          cleanup()
          reject(new Error('Streaming timeout'))
        }, 120000)

        runtime.EventsOn(`stream:chunk:${requestID}`, (chunk: string) => {
          if (chunk) {
            resultText = chunk
            setMiniModeResult(chunk)
          }
        })

        runtime.EventsOn(`stream:done:${requestID}`, () => {
          clearTimeout(timeout)
          cleanup()
          resolve()
        })

        runtime.EventsOn(`stream:error:${requestID}`, (errMsg: string) => {
          clearTimeout(timeout)
          cleanup()
          reject(new Error(errMsg))
        })

        RewriteAPI.StreamRewriteWithFormatting(requestID, selectedText, style, true)
      })

      if (resultText) {
        await AppAPI.ApplyRewrite(resultText)
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
currentStyle={settings.default_style ?? ''}
onExpand={handleMiniModeExpand}
onClose={handleClose}
onStyleChange={(style) => {
const newSettings = { ...settings, default_style: style }
handleSaveSettings(newSettings)
}}
onQuickRewrite={() => handleMiniModeRewrite(settings.default_style ?? '')}
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
