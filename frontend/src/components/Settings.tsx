import { useState, useEffect } from 'react'
import { Config, TextTypeInfo } from '../App'
import * as runtime from '../../wailsjs/runtime'
import '../styles/Settings.css'
import logoImage from '../assets/logo.jpg'

interface OperationResult {
  success: boolean
  error?: string
}

interface SettingsProps {
  settings: Config
  onSave: (settings: Config) => void
  onCancel: () => void
  onLoadCustomPrompts: () => Promise<Record<string, Record<string, string>>>
  onSaveCustomPrompt: (style: string, textType: string, prompt: string) => Promise<OperationResult>
  onDeleteCustomPrompt: (style: string, textType: string) => Promise<OperationResult>
  onGetDefaultPrompt: (style: string, textType: string) => Promise<string>
  onLoadRewriteStyles: () => Promise<string[]>
  onLoadAnalysisStyles: () => Promise<string[]>
  onLoadTextTypes: () => Promise<TextTypeInfo[]>
}

interface StyleOption {
  value: string
  label: string
  icon: string
}

function Settings({
  settings,
  onSave,
  onCancel,
  onLoadCustomPrompts,
  onSaveCustomPrompt,
  onDeleteCustomPrompt,
  onGetDefaultPrompt,
  onLoadRewriteStyles,
  onLoadAnalysisStyles,
  onLoadTextTypes
}: SettingsProps) {
  const [formData, setFormData] = useState<Config>(settings)
  const [availableModels, setAvailableModels] = useState<string[]>([])
  const [testingConnection, setTestingConnection] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'idle' | 'success' | 'error'>('idle')
  const [detectedVersion, setDetectedVersion] = useState('')
  const [saving, setSaving] = useState(false)

  const [customPrompts, setCustomPrompts] = useState<Record<string, Record<string, string>>>({})
  const [rewriteStyles, setRewriteStyles] = useState<StyleOption[]>([])
  const [analysisStyles, setAnalysisStyles] = useState<StyleOption[]>([])
  const [textTypes, setTextTypes] = useState<TextTypeInfo[]>([])
  const [selectedStyle, setSelectedStyle] = useState('')
  const [selectedTextType, setSelectedTextType] = useState('')
  const [customPromptText, setCustomPromptText] = useState('')
  const [hasCustomPrompt, setHasCustomPrompt] = useState(false)
  const [loadingPrompt, setLoadingPrompt] = useState(false)

  // Feedback state for user operations
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(null)
  const showFeedback = (type: 'success' | 'error', message: string) => {
    setFeedback({ type, message })
    setTimeout(() => setFeedback(null), 3000)
  }

  useEffect(() => {
    loadAvailableModels()
    loadSettingsData()
    runtime.WindowSetSize(580, 820)
    runtime.WindowCenter()
    runtime.WindowShow()
  }, [])

  const loadAvailableModels = async () => {
    try {
      // @ts-ignore
      const models = await window.go.main.App.GetAvailableModels()
      setAvailableModels(models || [])
    } catch (error) {
      console.error('Failed to load models:', error)
      setAvailableModels([])
    }
  }

  const loadSettingsData = async () => {
    try {
      const [prompts, rwStyles, anStyles, types] = await Promise.all([
        onLoadCustomPrompts(),
        onLoadRewriteStyles(),
        onLoadAnalysisStyles(),
        onLoadTextTypes()
      ])

      setCustomPrompts(prompts || {})

      const styleInfoPromises = rwStyles.map(style =>
        // @ts-ignore
        window.go.main.App.GetStyleInfo(style).catch(() => null)
      )
      const rwStyleInfos = await Promise.all(styleInfoPromises)

      setRewriteStyles(rwStyles.map((style, i) => ({
        value: style,
        label: rwStyleInfos[i]?.label || style,
        icon: rwStyleInfos[i]?.icon || '📝'
      })))

      const anStyleInfoPromises = anStyles.map(style =>
        // @ts-ignore
        window.go.main.App.GetStyleInfo(style).catch(() => null)
      )
      const anStyleInfos = await Promise.all(anStyleInfoPromises)

      setAnalysisStyles(anStyles.map((style, i) => ({
        value: style,
        label: anStyleInfos[i]?.label || style,
        icon: anStyleInfos[i]?.icon || '📋'
      })))

      setTextTypes(types || [])

      if (rwStyles.length > 0) {
        setSelectedStyle(rwStyles[0])
      }
      if (types.length > 0) {
        setSelectedTextType(types[0].type)
      }
    } catch (error) {
      console.error('Failed to load settings data:', error)
    }
  }

  const loadPromptForSelection = async () => {
    if (!selectedStyle || !selectedTextType) return

    setLoadingPrompt(true)
    try {
      const key = `${selectedStyle}.${selectedTextType}`
      const custom = customPrompts[selectedStyle]?.[selectedTextType]
      const isCustom = custom !== undefined && custom !== ''

      setHasCustomPrompt(isCustom)
      setCustomPromptText(isCustom ? custom : '')
    } finally {
      setLoadingPrompt(false)
    }
  }

  useEffect(() => {
    loadPromptForSelection()
  }, [selectedStyle, selectedTextType, customPrompts])

  const handleSaveCustomPrompt = async () => {
    if (!selectedStyle || !selectedTextType) return

    const result = await onSaveCustomPrompt(selectedStyle, selectedTextType, customPromptText)

    if (result.success) {
      const updated = { ...customPrompts }
      if (!updated[selectedStyle]) {
        updated[selectedStyle] = {}
      }
      updated[selectedStyle][selectedTextType] = customPromptText
      setCustomPrompts(updated)
      setHasCustomPrompt(true)
      showFeedback('success', 'Custom prompt saved successfully')
    } else {
      showFeedback('error', result.error || 'Failed to save custom prompt')
    }
  }

  const handleDeleteCustomPrompt = async () => {
    if (!selectedStyle || !selectedTextType) return

    const result = await onDeleteCustomPrompt(selectedStyle, selectedTextType)

    if (result.success) {
      const updated = { ...customPrompts }
      if (updated[selectedStyle]) {
        delete updated[selectedStyle][selectedTextType]
        if (Object.keys(updated[selectedStyle]).length === 0) {
          delete updated[selectedStyle]
        }
      }
      setCustomPrompts(updated)
      showFeedback('success', 'Custom prompt reset to default')
    } else {
      showFeedback('error', result.error || 'Failed to reset custom prompt')
    }

    const defaultPrompt = await onGetDefaultPrompt(selectedStyle, selectedTextType)
    setCustomPromptText(defaultPrompt)
    setHasCustomPrompt(false)
  }

  const handleLoadDefault = async () => {
    if (!selectedStyle || !selectedTextType) return
    const defaultPrompt = await onGetDefaultPrompt(selectedStyle, selectedTextType)
    setCustomPromptText(defaultPrompt)
  }

  const testConnection = async () => {
    setTestingConnection(true)
    setConnectionStatus('idle')
    setDetectedVersion('')

    try {
      // @ts-ignore
      const version = await window.go.main.App.TestConnection(
        formData.server_url,
        formData.model,
        formData.api_key || ""
      )
      setDetectedVersion(version)
      setConnectionStatus('success')
    } catch (error) {
      setConnectionStatus('error')
    }

    setTestingConnection(false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    await onSave(formData)
    setSaving(false)
  }

  const handleChange = (field: keyof Config, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  return (
    <div className="settings">
      <div className="settings-header">
        <div className="logo-container">
          <img src={logoImage} alt="App Logo" className="app-logo" />
        </div>
        <h2><span>⚙️</span> Settings</h2>
      </div>

      <form onSubmit={handleSubmit} className="settings-form">
        <div className="settings-section">
          <h3>Ollama Node</h3>

          <div className="form-group">
            <label htmlFor="server_url">Endpoint URL</label>
            <input
              type="text"
              id="server_url"
              value={formData.server_url}
              onChange={(e) => handleChange('server_url', e.target.value)}
              placeholder="http://localhost:11434"
            />
            <small>Default local node usually runs on port 11434</small>
          </div>

          <div className="form-group">
            <label htmlFor="model">Model Identifier</label>
            <div className="model-input-group">
              <input
                type="text"
                id="model"
                value={formData.model}
                onChange={(e) => handleChange('model', e.target.value)}
                placeholder="gemma3:1b"
                list="available-models"
              />
              <datalist id="available-models">
                {availableModels.map(model => (
                  <option key={model} value={model} />
                ))}
              </datalist>
            </div>
          </div>

          <button
            type="button"
            className="test-btn"
            onClick={testConnection}
            disabled={testingConnection}
          >
            {testingConnection ? 'Verifying node...' : 'Verify Connection'}
          </button>

          {connectionStatus !== 'idle' && (
            <div className={`status-message ${connectionStatus}`}>
              {connectionStatus === 'success'
                ? `✓ Node connected (Version: ${detectedVersion || 'Unknown'})`
                : '✗ Failed to connect to node'}
            </div>
          )}
        </div>

        <div className="settings-section">
          <h3>Preferences</h3>

          <div className="form-group">
            <label className="checkbox-label">
              <span>Start with Windows</span>
              <input
                type="checkbox"
                checked={formData.auto_start}
                onChange={(e) => handleChange('auto_start', e.target.checked)}
              />
              <div className="switch"></div>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <span>Clipboard Monitoring</span>
              <input
                type="checkbox"
                checked={formData.monitor_clipboard}
                onChange={(e) => handleChange('monitor_clipboard', e.target.checked)}
              />
              <div className="switch"></div>
            </label>
            <small>Automatically detect text when copied to clipboard</small>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <span>Use Compact Mode</span>
              <input
                type="checkbox"
                checked={formData.mini_mode || false}
                onChange={(e) => handleChange('mini_mode', e.target.checked)}
              />
              <div className="switch"></div>
            </label>
            <small>Show compact floating widget instead of full popup</small>
          </div>

          <div className="form-group">
            <label htmlFor="hotkey">Global Activation Key</label>
            <input
              type="text"
              id="hotkey"
              value={formData.hotkey}
              onChange={(e) => handleChange('hotkey', e.target.value)}
              placeholder="ctrl+shift+r"
            />
            <small>Current shortcut for triggering the rewrite flow</small>
          </div>

          <div className="form-group">
            <label htmlFor="default_style">Initial Style</label>
            <select
              id="default_style"
              value={formData.default_style}
              onChange={(e) => handleChange('default_style', e.target.value)}
            >
              <option value="grammar">🛡️ Grammar & Spelling</option>
              <option value="paraphrase">🔄 Paraphrase (Standard)</option>
              <option value="standard">📝 Standard</option>
              <option value="formal">📢 Formal</option>
              <option value="casual">💬 Casual</option>
              <option value="creative">✨ Creative</option>
              <option value="short">📏 Short</option>
              <option value="expand">📖 Expand</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="auto_paste_mode">Auto-Paste Behavior</label>
            <select
              id="auto_paste_mode"
              value={formData.auto_paste_mode || 'ask'}
              onChange={(e) => handleChange('auto_paste_mode', e.target.value)}
            >
              <option value="ask">❓ Ask every time</option>
              <option value="always">✓ Always paste automatically</option>
              <option value="never">✗ Never paste (copy only)</option>
            </select>
            <small>What happens when you click "Replace Selection"</small>
          </div>

          <div className="form-group">
            <label htmlFor="popup_position_mode">Popup Position</label>
            <select
              id="popup_position_mode"
              value={formData.popup_position_mode || 'cursor'}
              onChange={(e) => handleChange('popup_position_mode', e.target.value)}
            >
              <option value="cursor">📍 Near cursor</option>
              <option value="center">🖥️ Screen center</option>
            </select>
            <small>Where the popup window appears</small>
          </div>
        </div>

        <div className="settings-section">
          <h3>Custom Prompts</h3>
          <p className="section-description">
            Customize prompts for specific styles and text types. Leave empty to use default prompts.
          </p>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="prompt_style">Style</label>
              <select
                id="prompt_style"
                value={selectedStyle}
                onChange={(e) => setSelectedStyle(e.target.value)}
              >
                <optgroup label="Rewrite Styles">
                  {rewriteStyles.map(style => (
                    <option key={style.value} value={style.value}>
                      {style.icon} {style.label}
                    </option>
                  ))}
                </optgroup>
                <optgroup label="Analysis Styles">
                  {analysisStyles.map(style => (
                    <option key={style.value} value={style.value}>
                      {style.icon} {style.label}
                    </option>
                  ))}
                </optgroup>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="prompt_text_type">Text Type</label>
              <select
                id="prompt_text_type"
                value={selectedTextType}
                onChange={(e) => setSelectedTextType(e.target.value)}
              >
                {textTypes.map(type => (
                  <option key={type.type} value={type.type}>
                    {type.icon} {type.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="custom_prompt">
              Custom Prompt
              {hasCustomPrompt && <span className="custom-badge">Custom</span>}
            </label>
            <textarea
              id="custom_prompt"
              value={customPromptText}
              onChange={(e) => setCustomPromptText(e.target.value)}
              placeholder="Enter custom prompt... Leave empty to use default."
              rows={8}
              className="prompt-textarea"
            />
          </div>

          <div className="prompt-actions">
            <button
              type="button"
              className="secondary-btn"
              onClick={handleLoadDefault}
              disabled={loadingPrompt}
            >
              Load Default
            </button>
            <button
              type="button"
              className="delete-btn"
              onClick={handleDeleteCustomPrompt}
              disabled={!hasCustomPrompt || loadingPrompt}
            >
              Reset to Default
            </button>
            <button
              type="button"
              className="save-prompt-btn"
              onClick={handleSaveCustomPrompt}
              disabled={loadingPrompt}
            >
              Save Prompt
            </button>
          </div>
        </div>

        <div className="settings-footer">
          <button
            type="button"
            className="cancel-btn"
            onClick={onCancel}
          >
            Go Back
          </button>
          <button
            type="submit"
            className="save-btn"
            disabled={saving}
          >
            {saving ? 'Applying...' : 'Apply Changes'}
          </button>
        </div>
      </form>

      {feedback && (
        <div className={`feedback-toast ${feedback.type}`}>
          {feedback.message}
        </div>
      )}
    </div>
  )
}

export default Settings
