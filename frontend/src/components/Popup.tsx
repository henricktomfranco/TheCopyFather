import { useState, useEffect, useCallback, useRef } from 'react'
import * as runtime from '../../wailsjs/runtime'
import '../styles/Popup.css'

interface TextTypeInfo {
  type: string
  label: string
  icon: string
  description: string
}

interface DetectedTextType {
  type: string
  label: string
  icon: string
  confidence: number
}

interface PopupProps {
  originalText: string
  onSelect: (text: string) => void
  onClose: () => void
  onSettings: () => void
  defaultStyle?: string
  miniModeResult?: string
}

const PARAPHRASE_STYLES = [
  { value: 'standard', label: 'Standard', icon: '📝', desc: 'Balanced and natural rewrite' },
  { value: 'formal', label: 'Formal', icon: '📢', desc: 'Professional and academic tone' },
  { value: 'casual', label: 'Casual', icon: '💬', desc: 'Friendly and conversational' },
  { value: 'creative', label: 'Creative', icon: '✨', desc: 'Expressive and engaging' },
  { value: 'short', label: 'Short', icon: '📏', desc: 'Concise and clear' },
  { value: 'expand', label: 'Expand', icon: '📖', desc: 'Detailed and informative' },
]

const ANALYSIS_STYLES = [
  { value: 'summarize', label: 'TL;DR', icon: '📋', desc: 'Concise summary' },
  { value: 'bullets', label: 'Key Points', icon: '•••', desc: 'Bullet list of main points' },
  { value: 'insights', label: 'Insights', icon: '💡', desc: 'Key facts and arguments' },
]

function Popup({
  originalText,
  onSelect,
  onClose,
  onSettings,
  defaultStyle = 'grammar',
  miniModeResult
}: PopupProps) {
  const isGrammarDefault = defaultStyle === 'grammar'
  const initialMode = isGrammarDefault ? 'rewrite' : 'rewrite'
  const initialRewriteStyle = isGrammarDefault ? 'grammar' : (PARAPHRASE_STYLES.find(s => s.value === defaultStyle)?.value || 'standard')
  const initialAnalysisStyle = 'summarize'

  const [mainMode, setMainMode] = useState<'rewrite' | 'analyze'>(initialMode)
  const [rewriteStyle, setRewriteStyle] = useState(initialRewriteStyle)
  const [analysisStyle, setAnalysisStyle] = useState(initialAnalysisStyle)
  const [result, setResult] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [dropdownOpen, setDropdownOpen] = useState(false)

  // Auto-paste dialog state
  const [showPasteDialog, setShowPasteDialog] = useState(false)
  const [dontAskAgain, setDontAskAgain] = useState(false)
  const [autoPasteMode, setAutoPasteMode] = useState<string>('ask')

  // Formatting toggle state
  const [enableFormatting, setEnableFormatting] = useState(true)
  const enableFormattingRef = useRef(enableFormatting)

  // Keep ref in sync with state
  useEffect(() => {
    enableFormattingRef.current = enableFormatting
  }, [enableFormatting])

  // Result cache for lazy loading optimization
  const styleCacheRef = useRef<Map<string, string>>(new Map())

  // Confidence score state
  const [confidenceScore, setConfidenceScore] = useState<number | null>(null)

  // Undo/Redo stack state
  const [resultHistory, setResultHistory] = useState<Array<{ text: string; style: string; timestamp: number }>>([])
  const [historyIndex, setHistoryIndex] = useState<number>(-1)
  const MAX_HISTORY = 20

  // Text type detection state
  const [detectedTextType, setDetectedTextType] = useState<DetectedTextType | null>(null)
  const [selectedTextType, setSelectedTextType] = useState<string>('')
  const [availableTextTypes, setAvailableTextTypes] = useState<TextTypeInfo[]>([])
  const [textTypeDropdownOpen, setTextTypeDropdownOpen] = useState(false)
  const [isDetecting, setIsDetecting] = useState(false)
  const [isUserOverride, setIsUserOverride] = useState(false) // Track if user manually changed type

  const dropdownRef = useRef<HTMLDivElement>(null)
  const textTypeDropdownRef = useRef<HTMLDivElement>(null)

  // Load settings on mount
  useEffect(() => {
    const loadSettings = async () => {
      try {
        // @ts-ignore
        const settings = await window.go.main.App.GetSettings()
        if (settings) {
          setAutoPasteMode(settings.auto_paste_mode || 'ask')
        }
      } catch (e) {
        console.error('Failed to load settings:', e)
      }
    }
    loadSettings()
  }, [])

  // Load available text types and detect text type
  useEffect(() => {
    const loadTextTypesAndDetect = async () => {
      try {
        // Load available text types
        // @ts-ignore
        const types = await window.go.main.App.GetTextTypes()
        setAvailableTextTypes(types)

        // Detect text type if we have text
        if (originalText) {
          setIsDetecting(true)
          // @ts-ignore
          const detected = await window.go.main.App.DetectTextType(originalText)
          setDetectedTextType(detected)
          setSelectedTextType(detected.type)
          setIsDetecting(false)
        }
      } catch (e) {
        console.error('Failed to load text types:', e)
        setIsDetecting(false)
      }
    }
    loadTextTypesAndDetect()
  }, [originalText])

  // Handle outside click for text type dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (textTypeDropdownRef.current && !textTypeDropdownRef.current.contains(event.target as Node)) {
        setTextTypeDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const generate = useCallback(async (targetMainMode: string, targetStyle: string, useTextType: boolean = false) => {
    if (!originalText) return

    // Use ref for latest values to avoid stale closure issues
    const currentFormatting = enableFormattingRef.current
    const currentTextType = selectedTextType
    const cache = styleCacheRef.current

    // Create cache key based on mode, style, whether using text type, and formatting preference
    const cacheKey = `${targetMainMode}-${targetStyle}-${useTextType ? currentTextType : 'auto'}-${currentFormatting}`

    console.log('Generate called:', {
      targetMainMode, 
      targetStyle, 
      useTextType, 
      selectedTextType: currentTextType, 
      cacheKey,
      isUserOverride
    })

    // Check cache first
    if (cache.has(cacheKey)) {
      console.log('Using cached result for key:', cacheKey)
      setResult(cache.get(cacheKey)!)
      setError(null)
      return
    }

    setLoading(true)
    setError(null)
    setResult('')

    try {
      let generatedText = ''

      // Use text type if provided (either detected or user overridden)
      const textTypeToUse = useTextType ? currentTextType : 'normal'
      const useTypeSpecific = useTextType

      console.log('Using text type specific:', useTypeSpecific, 'Text type:', textTypeToUse)

      if (targetMainMode === 'analyze') {
        if (useTypeSpecific) {
          // @ts-ignore
          const option = await window.go.main.App.RetryAnalysisWithTextType(originalText, targetStyle, textTypeToUse, currentFormatting)
          console.log('Analysis with text type result:', option)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        } else {
          // Use generic analysis without text type
          // @ts-ignore
          const option = await window.go.main.App.RetryAnalysisWithFormatting(originalText, targetStyle, currentFormatting)
          console.log('Analysis generic result:', option)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        }
      } else {
        // Rewrite mode
        if (useTypeSpecific) {
          // @ts-ignore
          const option = await window.go.main.App.RetryRewriteWithTextType(originalText, targetStyle, textTypeToUse, currentFormatting)
          console.log('Rewrite with text type result:', option)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        } else {
          // Use generic rewrite without text type
          // @ts-ignore
          const option = await window.go.main.App.RetryRewriteWithFormatting(originalText, targetStyle, currentFormatting)
          console.log('Rewrite generic result:', option)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        }
      }

      // Cache the successful result and add to history
      if (generatedText) {
        cache.set(cacheKey, generatedText)
        
        // Add to history
        const historyEntry = {
          text: generatedText,
          style: targetStyle,
          timestamp: Date.now()
        }
        setResultHistory(prev => {
          // Remove any entries after current index (for redo support)
          const newHistory = prev.slice(0, historyIndex + 1)
          newHistory.push(historyEntry)
          // Keep only MAX_HISTORY entries
          if (newHistory.length > MAX_HISTORY) {
            newHistory.shift()
          }
          return newHistory
        })
        setHistoryIndex(prev => Math.min(prev + 1, MAX_HISTORY - 1))
        
        // Calculate confidence score (mock for now - in real implementation this would come from the backend)
        // Higher confidence for grammar style, lower for creative
        const baseConfidence = targetStyle === 'grammar' ? 0.92 : 
                              targetStyle === 'formal' ? 0.88 :
                              targetStyle === 'casual' ? 0.85 :
                              targetStyle === 'creative' ? 0.75 : 0.82
        // Add some randomness
        const confidence = Math.min(0.98, Math.max(0.60, baseConfidence + (Math.random() * 0.1 - 0.05)))
        setConfidenceScore(Math.round(confidence * 100))
      }
    } catch (err) {
      console.error('Generate error:', err)
      setError('Failed to connect to AI server')
    }
    setLoading(false)
  }, [originalText, selectedTextType, isUserOverride, historyIndex])

  // Initial generation after text type detection completes
  useEffect(() => {
    if (!originalText) return
    
    // Resize window for popup mode
    runtime.WindowSetSize(500, 600)
    runtime.WindowSetAlwaysOnTop(true)
    runtime.WindowShow()
    
    // If we have a result from mini mode, use it
    if (miniModeResult) {
      setResult(miniModeResult)
      setConfidenceScore(85) // Default confidence for mini mode results
      setResultHistory([{ text: miniModeResult, style: rewriteStyle, timestamp: Date.now() }])
      setHistoryIndex(0)
      return
    }
    
    // Only generate after text type detection is complete (not loading)
    if (!isDetecting && detectedTextType) {
      // Use detected text type if confidence is high enough (>= 0.6)
      const shouldUseDetectedType = detectedTextType.confidence >= 0.6
      generate(initialMode, isGrammarDefault ? 'grammar' : initialRewriteStyle, shouldUseDetectedType)
    } else if (!isDetecting) {
      // Detection completed but no result, use normal
      generate(initialMode, isGrammarDefault ? 'grammar' : initialRewriteStyle, false)
    }
    // If still detecting, wait for it to complete
  }, [generate, initialMode, isGrammarDefault, initialRewriteStyle, isDetecting, detectedTextType, originalText, miniModeResult, rewriteStyle])

  // Handle outside click for dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Helper to determine if we should use detected text type
  const shouldUseTextType = () => {
    return detectedTextType !== null && detectedTextType.confidence >= 0.6
  }

  const handleMainModeChange = (newMode: 'rewrite' | 'analyze') => {
    if (newMode === mainMode) return
    setMainMode(newMode)
    const useTextType = shouldUseTextType()
    if (newMode === 'analyze') {
      generate('analyze', analysisStyle, useTextType)
    } else {
      generate('rewrite', rewriteStyle, useTextType)
    }
  }

  const handleRewriteStyleChange = (newStyle: string) => {
    setRewriteStyle(newStyle)
    setDropdownOpen(false)
    generate('rewrite', newStyle, shouldUseTextType())
  }

  const handleAnalysisStyleChange = (newStyle: string) => {
    setAnalysisStyle(newStyle)
    setDropdownOpen(false)
    generate('analyze', newStyle, shouldUseTextType())
  }

  const handleTextTypeChange = (newType: string) => {
    setSelectedTextType(newType)
    setTextTypeDropdownOpen(false)
    setIsUserOverride(true)
    // Regenerate with the new text type (user override)
    if (mainMode === 'analyze') {
      generate('analyze', analysisStyle, true)
    } else {
      generate('rewrite', rewriteStyle, true)
    }
  }

  // Undo/Redo functions
  const handleUndo = useCallback(() => {
    if (historyIndex > 0) {
      const newIndex = historyIndex - 1
      setHistoryIndex(newIndex)
      setResult(resultHistory[newIndex].text)
    }
  }, [historyIndex, resultHistory])

  const handleRedo = useCallback(() => {
    if (historyIndex < resultHistory.length - 1) {
      const newIndex = historyIndex + 1
      setHistoryIndex(newIndex)
      setResult(resultHistory[newIndex].text)
    }
  }, [historyIndex, resultHistory])

  // Keyboard shortcuts for undo/redo
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
        e.preventDefault()
        handleUndo()
      } else if ((e.ctrlKey || e.metaKey) && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) {
        e.preventDefault()
        handleRedo()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleUndo, handleRedo])

  const handleCopy = async () => {
    if (!result) return
    try {
      // Use the backend clipboard manager instead of browser clipboard
      // @ts-ignore
      await window.go.main.App.ApplyRewrite(result)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
      // Fallback to browser clipboard
      try {
        await navigator.clipboard.writeText(result)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      } catch (e) {
        console.error('Fallback copy also failed:', e)
      }
    }
  }

  const handleReplace = async () => {
    if (!result) return

    if (autoPasteMode === 'ask') {
      // Show the paste confirmation dialog
      setShowPasteDialog(true)
    } else if (autoPasteMode === 'always') {
      // Auto-paste directly
      try {
        // @ts-ignore
        await window.go.main.App.ApplyRewriteAndPaste(result)
        onClose()
      } catch (err) {
        console.error('Failed to paste:', err)
        // Fallback to just copying
        await handleCopy()
        onClose()
      }
    } else {
      // Never auto-paste, just copy and close
      await handleCopy()
      onClose()
    }
  }

  const handlePasteConfirm = async () => {
    if (dontAskAgain) {
      // Save the preference
      try {
        // @ts-ignore
        const settings = await window.go.main.App.GetSettings()
        settings.auto_paste_mode = 'always'
        // @ts-ignore
        await window.go.main.App.SaveSettings(settings)
        setAutoPasteMode('always')
      } catch (e) {
        console.error('Failed to save settings:', e)
      }
    }

    setShowPasteDialog(false)

    try {
      // @ts-ignore
      await window.go.main.App.ApplyRewriteAndPaste(result)
      onClose()
    } catch (err) {
      console.error('Failed to paste:', err)
      // Fallback to just copying
      await handleCopy()
      onClose()
    }
  }

  const handlePasteCancel = async () => {
    if (dontAskAgain) {
      // Save the preference to never auto-paste
      try {
        // @ts-ignore
        const settings = await window.go.main.App.GetSettings()
        settings.auto_paste_mode = 'never'
        // @ts-ignore
        await window.go.main.App.SaveSettings(settings)
        setAutoPasteMode('never')
      } catch (e) {
        console.error('Failed to save settings:', e)
      }
    }

    setShowPasteDialog(false)
    // Just copy, don't paste
    await handleCopy()
    onClose()
  }

  const currentRewriteStyleData = PARAPHRASE_STYLES.find(s => s.value === rewriteStyle) || PARAPHRASE_STYLES[0]
  const currentAnalysisStyleData = ANALYSIS_STYLES.find(s => s.value === analysisStyle) || ANALYSIS_STYLES[0]

  const renderContent = (text: string) => {
    if (!text) return null
    console.log('Rendering text:', text.substring(0, 100) + '...')

    // Clean up any double bolding like ****
    let cleanText = text.replace(/\*\*\*\*/g, '**')

    // Also handle single asterisk bold (*text*)
    cleanText = cleanText.replace(/\*([^\s*][^*]*[^\s*])\*/g, '**$1**')

    console.log('Cleaned text has bold markers:', cleanText.includes('**'))

    // Check if this is a bullet list (for bullets analysis style)
    const lines = cleanText.split('\n')
    const isBulletList = lines.some(line => line.trim().match(/^[-•\*]\s/))

    if (isBulletList && mainMode === 'analyze' && analysisStyle === 'bullets') {
      // Render as actual bullet list
      return (
        <ul className="bullet-list">
          {lines.map((line, i) => {
            const trimmed = line.trim()
            if (!trimmed) return null
            // Remove bullet markers
            const content = trimmed.replace(/^[-•\*]\s*/, '')
            // Handle bold text - match **text** patterns
            const parts = content.split(/(\*\*[^*]+\*\*)/g)
            return (
              <li key={i}>
                {parts.map((part, j) => {
                  if (part && part.startsWith('**') && part.endsWith('**')) {
                    return <strong key={j} className="highlight-bold">{part.slice(2, -2)}</strong>
                  }
                  return part
                })}
              </li>
            )
          })}
        </ul>
      )
    }

    // Check if this looks like an email (has greeting and signature)
    const isEmail = cleanText.match(/^(Dear\s|Hi\s|Hello\s|To\s)/i) &&
      cleanText.match(/(Regards|Sincerely|Thanks|Best)/i)

    if (isEmail) {
      // Format email with better structure
      return cleanText.split(/\n\n+/).map((para, i) => {
        // Match **text** patterns
        const parts = para.split(/(\*\*[^*]+\*\*)/g)
        const isGreeting = para.match(/^(Dear\s|Hi\s|Hello\s|To\s)/i)
        const isClosing = para.match(/(Regards|Sincerely|Thanks|Best)/i) && para.length < 100

        return (
          <p
            key={i}
            className={`${isGreeting ? 'email-greeting' : ''} ${isClosing ? 'email-closing' : ''}`}
          >
            {parts.map((part, j) => {
              if (part && part.startsWith('**') && part.endsWith('**')) {
                return <strong key={j} className="highlight-bold">{part.slice(2, -2)}</strong>
              }
              return part
            })}
          </p>
        )
      })
    }

    // Split by paragraphs for normal content
    return cleanText.split(/\n\n+/).map((para, i) => {
      // Match **text** patterns more reliably
      const parts = para.split(/(\*\*[^*]+\*\*)/g)
      return (
        <p key={i}>
          {parts.map((part, j) => {
            if (part && part.startsWith('**') && part.endsWith('**')) {
              return <strong key={j} className="highlight-bold">{part.slice(2, -2)}</strong>
            }
            return part
          })}
        </p>
      )
    })
  }

  return (
    <div className="popup modern">
      <div className="popup-nav">
        <div className="mode-toggle">
          <button
            className={`toggle-btn ${mainMode === 'rewrite' ? 'active' : ''}`}
            onClick={() => handleMainModeChange('rewrite')}
          >
            🔄 Rewrite
          </button>
          <button
            className={`toggle-btn ${mainMode === 'analyze' ? 'active' : ''}`}
            onClick={() => handleMainModeChange('analyze')}
          >
            📊 Analyze
          </button>
        </div>
        <div className="formatting-toggle-container">
          <label className="formatting-toggle" title={enableFormatting ? "Formatting enabled" : "Plain text mode"}>
            <input
              type="checkbox"
              checked={enableFormatting}
              onChange={(e) => {
                const newValue = e.target.checked
                setEnableFormatting(newValue)
                // Regenerate with new formatting preference after state updates
                setTimeout(() => {
                  if (mainMode === 'analyze') {
                    generate('analyze', analysisStyle, shouldUseTextType())
                  } else {
                    generate('rewrite', rewriteStyle, shouldUseTextType())
                  }
                }, 0)
              }}
            />
            <span className="toggle-slider"></span>
            <span className="toggle-label">{enableFormatting ? '✨' : 'T'}</span>
          </label>
        </div>
        <div className="nav-actions">
          <button className="icon-btn" onClick={onSettings} title="Settings">⚙️</button>
          <button className="icon-btn" onClick={onClose} title="Close">✕</button>
        </div>
      </div>


      <div className="content-area">
        {/* Text Type Detection Badge */}
        {selectedTextType && (
          <div className="text-type-section" ref={textTypeDropdownRef}>
            <div className="text-type-label">
              {isUserOverride ? 'Applied Type:' : 'Detected Type:'}
            </div>
            <div className="custom-dropdown text-type-dropdown">
              <div
                className={`dropdown-trigger text-type-trigger ${textTypeDropdownOpen ? 'open' : ''} ${!isUserOverride && detectedTextType && detectedTextType.confidence < 0.6 ? 'low-confidence' : ''} ${isUserOverride ? 'user-override' : ''}`}
                onClick={() => setTextTypeDropdownOpen(!textTypeDropdownOpen)}
                title={isUserOverride 
                  ? 'You selected this type - rewrite uses type-specific prompts' 
                  : detectedTextType && detectedTextType.confidence < 0.6 
                    ? 'Low confidence - click to correct' 
                    : 'Auto-detected (click to apply type-specific rewrite)'}
              >
                <span>
                  {isDetecting ? (
                    <>🔍 Analyzing...</>
                  ) : (
                    <>
                      {availableTextTypes.find(t => t.type === selectedTextType)?.icon || '📝'} {' '}
                      {availableTextTypes.find(t => t.type === selectedTextType)?.label || 'Text'}
                      {!isUserOverride && detectedTextType && detectedTextType.confidence < 0.6 && ' ⚠️'}
                      {isUserOverride && ' ✓'}
                    </>
                  )}
                </span>
                <span className="chevron">↓</span>
              </div>

              {textTypeDropdownOpen && (
                <div className="dropdown-menu text-type-menu">
                  <div className="dropdown-header">
                    <span>{isUserOverride ? 'Change applied type:' : 'Detected type - select to apply:'}</span>
                    {detectedTextType && !isUserOverride && (
                      <span className="confidence-hint">
                        Confidence: {(detectedTextType.confidence * 100).toFixed(0)}%
                      </span>
                    )}
                    {isUserOverride && (
                      <span className="override-hint">
                        Type-specific rewrite active
                      </span>
                    )}
                  </div>
                  <div className="dropdown-divider"></div>
                  {availableTextTypes.map(t => (
                    <div
                      key={t.type}
                      className={`dropdown-item ${t.type === selectedTextType ? 'active' : ''}`}
                      onClick={() => handleTextTypeChange(t.type)}
                    >
                      <div className="item-icon">{t.icon}</div>
                      <div className="item-content">
                        <div className="item-label">{t.label}</div>
                        <div className="item-desc">{t.description}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {mainMode === 'rewrite' && (
          <div className="style-selector" ref={dropdownRef}>
            <div className="custom-dropdown">
              <div
                className={`dropdown-trigger ${dropdownOpen ? 'open' : ''}`}
                onClick={() => setDropdownOpen(!dropdownOpen)}
              >
                <span>{currentRewriteStyleData.icon} {currentRewriteStyleData.label}</span>
                <span className="chevron">↓</span>
              </div>

              {dropdownOpen && (
                <div className="dropdown-menu">
                  <div
                    className={`dropdown-item ${rewriteStyle === 'grammar' ? 'active' : ''}`}
                    onClick={() => handleRewriteStyleChange('grammar')}
                  >
                    <div className="item-icon">🛡️</div>
                    <div className="item-content">
                      <div className="item-label">Grammar & Spelling</div>
                      <div className="item-desc">Corrects errors and improves flow</div>
                    </div>
                  </div>
                  <div className="dropdown-divider"></div>
                  {PARAPHRASE_STYLES.map(s => (
                    <div
                      key={s.value}
                      className={`dropdown-item ${s.value === rewriteStyle ? 'active' : ''}`}
                      onClick={() => handleRewriteStyleChange(s.value)}
                    >
                      <div className="item-icon">{s.icon}</div>
                      <div className="item-content">
                        <div className="item-label">{s.label}</div>
                        <div className="item-desc">{s.desc}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {mainMode === 'analyze' && (
          <div className="style-selector" ref={dropdownRef}>
            <div className="custom-dropdown">
              <div
                className={`dropdown-trigger ${dropdownOpen ? 'open' : ''}`}
                onClick={() => setDropdownOpen(!dropdownOpen)}
              >
                <span>{currentAnalysisStyleData.icon} {currentAnalysisStyleData.label}</span>
                <span className="chevron">↓</span>
              </div>

              {dropdownOpen && (
                <div className="dropdown-menu">
                  {ANALYSIS_STYLES.map(s => (
                    <div
                      key={s.value}
                      className={`dropdown-item ${s.value === analysisStyle ? 'active' : ''}`}
                      onClick={() => handleAnalysisStyleChange(s.value)}
                    >
                      <div className="item-icon">{s.icon}</div>
                      <div className="item-content">
                        <div className="item-label">{s.label}</div>
                        <div className="item-desc">{s.desc}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        <div className={`result-container ${loading ? 'loading' : ''}`}>
          <div className="result-header">
            <div className="result-header-left">
              {mainMode === 'rewrite' && rewriteStyle === 'grammar' && !loading && result && (
                <div className="grammar-badge">
                  ✓ Grammar & Style Improved
                </div>
              )}
              {!loading && confidenceScore !== null && result && (
                <div className={`confidence-badge ${confidenceScore >= 85 ? 'high' : confidenceScore >= 70 ? 'medium' : 'low'}`}>
                  <span className="confidence-icon">📊</span>
                  <span className="confidence-value">{confidenceScore}%</span>
                  <span className="confidence-label">confidence</span>
                </div>
              )}
            </div>
            <div className="result-header-center">
              {/* Undo/Redo Controls */}
              {resultHistory.length > 0 && (
                <div className="history-controls">
                  <button
                    className="history-btn"
                    onClick={handleUndo}
                    disabled={historyIndex <= 0}
                    title="Undo (Ctrl+Z)"
                  >
                    ↶
                  </button>
                  <span className="history-indicator">
                    {historyIndex + 1} / {resultHistory.length}
                  </span>
                  <button
                    className="history-btn"
                    onClick={handleRedo}
                    disabled={historyIndex >= resultHistory.length - 1}
                    title="Redo (Ctrl+Y or Ctrl+Shift+Z)"
                  >
                    ↷
                  </button>
                </div>
              )}
            </div>
            <div className="result-header-right">
              <div className="copy-hint">{copied ? '✓ COPIED' : 'CLICK TO REPLACE'}</div>
            </div>
          </div>
          {loading ? (
            <div className="skeleton-loader">
              <div className="line"></div>
              <div className="line"></div>
              <div className="line short"></div>
            </div>
          ) : error ? (
            <div className="error-box">
              <div className="error-icon">⚠️</div>
              <p>{error}</p>
              <button className="secondary-btn" onClick={() => generate(mainMode, mainMode === 'analyze' ? analysisStyle : rewriteStyle, shouldUseTextType())}>Retry Connection</button>
            </div>
          ) : (
            <div className="result-text markdown-content" onClick={() => onSelect(result)}>
              {renderContent(result)}
            </div>
          )}
        </div>
      </div>

      <div className="modern-footer">
        <button className="secondary-btn" onClick={handleCopy}>
          {copied ? '✓ Copied' : '📋 Copy'}
        </button>
        <button
          className="primary-btn"
          onClick={handleReplace}
          disabled={loading || !!error || !result}
        >
          Replace Selection
        </button>
      </div>

      {/* Paste Confirmation Dialog */}
      {showPasteDialog && (
        <div className="paste-dialog-overlay">
          <div className="paste-dialog">
            <div className="paste-dialog-header">
              <h3>📋 Replace Text?</h3>
              <p>Automatically paste the rewritten text?</p>
            </div>
            <div className="paste-dialog-content">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={dontAskAgain}
                  onChange={(e) => setDontAskAgain(e.target.checked)}
                />
                <span>Don't ask again</span>
              </label>
            </div>
            <div className="paste-dialog-footer">
              <button className="secondary-btn" onClick={handlePasteCancel}>
                Copy Only
              </button>
              <button className="primary-btn" onClick={handlePasteConfirm}>
                ✓ Paste
              </button>
            </div>
          </div>
        </div>
      )}

      <style>{`
        /* Formatting Toggle Styles */
        .formatting-toggle-container {
          display: flex;
          align-items: center;
          margin: 0 12px;
        }
        
        .formatting-toggle {
          position: relative;
          display: inline-flex;
          align-items: center;
          cursor: pointer;
          user-select: none;
        }
        
        .formatting-toggle input[type="checkbox"] {
          position: absolute;
          opacity: 0;
          width: 0;
          height: 0;
        }
        
        .formatting-toggle .toggle-slider {
          position: relative;
          display: inline-block;
          width: 40px;
          height: 22px;
          background: rgba(255, 255, 255, 0.1);
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 11px;
          transition: all 0.3s ease;
        }
        
        .formatting-toggle .toggle-slider::before {
          content: '';
          position: absolute;
          width: 16px;
          height: 16px;
          left: 2px;
          top: 2px;
          background: white;
          border-radius: 50%;
          transition: all 0.3s ease;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        }
        
        .formatting-toggle input:checked + .toggle-slider {
          background: linear-gradient(135deg, #3b82f6, #2563eb);
          border-color: #3b82f6;
        }
        
        .formatting-toggle input:checked + .toggle-slider::before {
          transform: translateX(18px);
        }
        
        .formatting-toggle .toggle-label {
          position: absolute;
          left: 50%;
          top: 50%;
          transform: translate(-50%, -50%);
          font-size: 12px;
          pointer-events: none;
          z-index: 1;
        }
        
        .formatting-toggle:hover .toggle-slider {
          border-color: rgba(255, 255, 255, 0.4);
        }
        
        /* Existing Styles */
        .dropdown-trigger .chevron {
          font-size: 10px;
          transition: transform 0.2s;
        }
        .dropdown-trigger.open .chevron {
          transform: rotate(180deg);
        }
        .dropdown-menu {
          position: absolute;
          top: calc(100% + 8px);
          left: 0;
          right: 0;
          background: var(--bg-tertiary);
          border: 1px solid var(--border-bright);
          border-radius: var(--radius-md);
          box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
          z-index: 1000;
          padding: 6px;
          animation: menuAppear 0.2s cubic-bezier(0.16, 1, 0.3, 1);
        }
        @keyframes menuAppear {
          from { opacity: 0; transform: translateY(-4px) scale(0.98); }
          to { opacity: 1; transform: translateY(0) scale(1); }
        }
        .dropdown-item {
          display: flex;
          gap: 12px;
          padding: 10px 12px;
          border-radius: var(--radius-sm);
          cursor: pointer;
          transition: var(--transition-smooth);
        }
        .dropdown-item:hover {
          background: rgba(255, 255, 255, 0.05);
        }
        .dropdown-item.active {
          background: rgba(59, 130, 246, 0.1);
        }
        .dropdown-divider {
          height: 1px;
          background: var(--border-dim);
          margin: 6px 0;
        }
        .item-icon { font-size: 16px; }
        .item-label { font-size: 13px; font-weight: 500; color: var(--text-primary); }
        .item-desc { font-size: 11px; color: var(--text-muted); margin-top: 2px; }
        
        .error-icon { font-size: 32px; margin-bottom: 8px; }
        
        .result-container.loading {
          opacity: 0.7;
        }
        
        .bullet-list {
          list-style: none;
          padding: 0;
          margin: 0;
        }
        .bullet-list li {
          position: relative;
          padding-left: 20px;
          margin-bottom: 12px;
          line-height: 1.5;
        }
        .bullet-list li::before {
          content: "•";
          position: absolute;
          left: 0;
          color: var(--accent-color, #3b82f6);
          font-weight: bold;
        }
        .bullet-list li:last-child {
          margin-bottom: 0;
        }
        
        .highlight-bold {
          background: linear-gradient(135deg, rgba(59, 130, 246, 0.25), rgba(59, 130, 246, 0.15));
          padding: 2px 6px;
          border-radius: 4px;
          font-weight: 700;
          color: #60a5fa;
          border: 1px solid rgba(59, 130, 246, 0.3);
          box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
        }
        
        .markdown-content strong {
          background: linear-gradient(135deg, rgba(59, 130, 246, 0.25), rgba(59, 130, 246, 0.15));
          padding: 2px 6px;
          border-radius: 4px;
          font-weight: 700;
          color: #60a5fa;
          border: 1px solid rgba(59, 130, 246, 0.3);
        }
        
        .email-greeting {
          margin-bottom: 16px;
          font-weight: 500;
        }
        
        .email-closing {
          margin-top: 20px;
          font-style: italic;
          color: var(--text-secondary, #a0a0a0);
        }
        
        .grammar-badge {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          background: rgba(34, 197, 94, 0.15);
          color: #22c55e;
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 11px;
          font-weight: 500;
        }
        
        .result-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 12px;
          flex-wrap: wrap;
          gap: 8px;
        }
        
        .result-header .copy-hint {
          margin: 0;
        }
        
        /* Paste Dialog Styles */
        .paste-dialog-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.7);
          backdrop-filter: blur(4px);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
          animation: fadeIn 0.2s ease;
        }
        
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        
        .paste-dialog {
          background: var(--bg-secondary, #1f1f1f);
          border: 1px solid var(--border-bright, #3a3a3a);
          border-radius: 12px;
          padding: 24px;
          width: 320px;
          max-width: 90%;
          box-shadow: 0 24px 48px rgba(0, 0, 0, 0.5);
          animation: dialogSlide 0.3s cubic-bezier(0.16, 1, 0.3, 1);
        }
        
        @keyframes dialogSlide {
          from { opacity: 0; transform: translateY(-20px) scale(0.95); }
          to { opacity: 1; transform: translateY(0) scale(1); }
        }
        
        .paste-dialog-header {
          text-align: center;
          margin-bottom: 20px;
        }
        
        .paste-dialog-header h3 {
          margin: 0 0 8px 0;
          font-size: 18px;
          font-weight: 600;
          color: var(--text-primary, #ffffff);
        }
        
        .paste-dialog-header p {
          margin: 0;
          font-size: 14px;
          color: var(--text-muted, #888888);
        }
        
        .paste-dialog-content {
          margin-bottom: 20px;
        }
        
        .checkbox-label {
          display: flex;
          align-items: center;
          gap: 10px;
          cursor: pointer;
          font-size: 14px;
          color: var(--text-secondary, #a0a0a0);
          user-select: none;
        }
        
        .checkbox-label input[type="checkbox"] {
          width: 18px;
          height: 18px;
          accent-color: #3b82f6;
          cursor: pointer;
        }
        
        .paste-dialog-footer {
          display: flex;
          gap: 12px;
          justify-content: center;
        }
        
        .paste-dialog-footer button {
          flex: 1;
          padding: 10px 16px;
          font-size: 14px;
        }

        /* Text Type Selector Styles */
        .text-type-section {
          display: flex;
          align-items: center;
          gap: 10px;
          margin-bottom: 16px;
          padding: 8px 12px;
          background: var(--bg-glass);
          backdrop-filter: var(--glass-backdrop);
          border-radius: var(--radius-md);
          border: 1px solid var(--border-subtle);
          position: relative;
          z-index: 100;
        }

        .text-type-label {
          font-size: 12px;
          color: var(--text-muted);
          white-space: nowrap;
        }

        .text-type-dropdown {
          flex: 1;
        }

        .text-type-trigger {
          font-size: 13px;
          padding: 6px 10px;
          background: rgba(255, 255, 255, 0.05);
          border-radius: var(--radius-sm);
          border: 1px solid var(--border-subtle);
        }

        .text-type-trigger:hover {
          background: rgba(255, 255, 255, 0.1);
          border-color: var(--border-bright);
        }

        .text-type-trigger.low-confidence {
          border-color: rgba(234, 179, 8, 0.5);
          background: rgba(234, 179, 8, 0.1);
        }

        .text-type-trigger.low-confidence:hover {
          border-color: rgba(234, 179, 8, 0.8);
          background: rgba(234, 179, 8, 0.15);
        }

        .text-type-menu {
          max-height: 300px;
          overflow-y: auto;
          z-index: 1000;
          position: relative;
        }

        .dropdown-header {
          padding: 8px 12px;
          font-size: 11px;
          color: var(--text-muted);
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .confidence-hint {
          font-size: 10px;
          color: var(--text-secondary);
          font-style: italic;
        }

        .override-hint {
          font-size: 10px;
          color: #22c55e;
          font-weight: 500;
        }

        .text-type-trigger.user-override {
          border-color: rgba(34, 197, 94, 0.5);
          background: rgba(34, 197, 94, 0.1);
        }

        .text-type-trigger.user-override:hover {
          border-color: rgba(34, 197, 94, 0.8);
          background: rgba(34, 197, 94, 0.15);
        }
      `}</style>
    </div>
  )
}

export default Popup
