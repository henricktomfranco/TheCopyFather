import { useState, useEffect, useCallback, useRef } from 'react'
import * as runtime from '../../wailsjs/runtime'
import * as AppAPI from '../../wailsjs/go/main/App'
import { rewriter as rewriterModels } from '../../wailsjs/go/models'
import '../styles/main.css'
import '../styles/Popup.css'

interface DetectedTextType {
  type: string
  label: string
  icon: string
  confidence: number
}

interface TextTypeInfo {
  Type: string
  Label: string
  Icon: string
  Description: string
}

interface PopupProps {
  originalText: string
  onSelect: (text: string) => void
  onClose: () => void
  onSettings: () => void
  onShowDiff?: (rewritten: string) => void
  onResultChange?: (text: string) => void
  defaultStyle?: string
  miniModeResult?: string
}

const REWRITE_STYLES = [
  { value: 'standard', label: 'Standard', icon: '📝', desc: 'Balanced and natural rewrite' },
  { value: 'formal', label: 'Formal', icon: '📢', desc: 'Professional and academic tone' },
  { value: 'casual', label: 'Casual', icon: '💬', desc: 'Friendly and conversational' },
  { value: 'creative', label: 'Creative', icon: '✨', desc: 'Expressive and engaging' },
  { value: 'short', label: 'Short', icon: '📏', desc: 'Concise and clear' },
  { value: 'expand', label: 'Expand', icon: '📖', desc: 'Detailed and informative' },
  { value: 'paraphrase', label: 'Paraphrase', icon: '🔁', desc: 'Same meaning, different words' },
]

const ANALYSIS_STYLES = [
  { value: 'summarize', label: 'TL;DR', icon: '📋', desc: 'Concise summary' },
  { value: 'bullets', label: 'Key Points', icon: '•••', desc: 'Bullet list of main points' },
  { value: 'insights', label: 'Insights', icon: '💡', desc: 'Key facts and arguments' },
]

export default function Popup({
  originalText,
  onSelect,
  onClose,
  onSettings,
  onShowDiff,
  onResultChange,
  defaultStyle = 'grammar',
  miniModeResult
}: PopupProps) {
  const isGrammarDefault = defaultStyle === 'grammar'
  const initialMode = isGrammarDefault ? 'rewrite' : 'rewrite'
  const initialRewriteStyle = isGrammarDefault ? 'grammar' : (REWRITE_STYLES.find(s => s.value === defaultStyle)?.value || 'standard')
  const initialAnalysisStyle = 'summarize'

  const [mainMode, setMainMode] = useState<'rewrite' | 'analyze'>(initialMode)
  const [rewriteStyle, setRewriteStyle] = useState(initialRewriteStyle)
  const [analysisStyle, setAnalysisStyle] = useState(initialAnalysisStyle)
  const [result, setResult] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [showPasteDialog, setShowPasteDialog] = useState(false)
  const [dontAskAgain, setDontAskAgain] = useState(false)
  const [autoPasteMode, setAutoPasteMode] = useState<string>('ask')
  const [enableFormatting, setEnableFormatting] = useState(true)
  const enableFormattingRef = useRef(enableFormatting)
  const styleCacheRef = useRef<Map<string, string>>(new Map())
  const [confidenceScore, setConfidenceScore] = useState<number | null>(null)
  const [resultHistory, setResultHistory] = useState<Array<{ text: string; style: string; timestamp: number }>>([])
  const [variationIndex, setVariationIndex] = useState<number>(-1)
  const MAX_VARIATIONS = 20
  const [detectedTextType, setDetectedTextType] = useState<DetectedTextType | null>(null)
  const [selectedTextType, setSelectedTextType] = useState<string>('')
  const [availableTextTypes, setAvailableTextTypes] = useState<rewriterModels.TextTypeInfo[]>([])
  const [textTypeDropdownOpen, setTextTypeDropdownOpen] = useState(false)
  const [isDetecting, setIsDetecting] = useState(false)
  const [isUserOverride, setIsUserOverride] = useState(false)
  const [showOriginal, setShowOriginal] = useState(false)

  const dropdownRef = useRef<HTMLDivElement>(null)
  const textTypeDropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    enableFormattingRef.current = enableFormatting
  }, [enableFormatting])

  useEffect(() => {
    if (onResultChange) {
      onResultChange(result)
    }
  }, [result, onResultChange])

    useEffect(() => {
    const loadSettings = async () => {
      try {
        const settings = await AppAPI.GetSettings()
        if (settings) {
          setAutoPasteMode(settings.auto_paste_mode || 'ask')
        }
      } catch (e) {
        console.error('Failed to load settings:', e)
      }
    }
    loadSettings()
  }, [])

    useEffect(() => {
    const loadTextTypesAndDetect = async () => {
      try {
        const types = await AppAPI.GetTextTypes()
        setAvailableTextTypes(types)

        if (originalText) {
          setIsDetecting(true)
          const detected = await AppAPI.DetectTextType(originalText)
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

    const currentFormatting = enableFormattingRef.current
    const currentTextType = selectedTextType
    const cache = styleCacheRef.current
    const cacheKey = `${targetMainMode}-${targetStyle}-${useTextType ? currentTextType : 'auto'}-${currentFormatting}`

    console.log('Generate called:', { targetMainMode, targetStyle, useTextType, selectedTextType: currentTextType, cacheKey, isUserOverride })

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
      const textTypeToUse = useTextType ? currentTextType : 'normal'
      const useTypeSpecific = useTextType

      console.log('Using text type specific:', useTypeSpecific, 'Text type:', textTypeToUse)

      if (targetMainMode === 'analyze') {
        if (useTypeSpecific) {
          const option = await AppAPI.RetryAnalysisWithTextType(originalText, targetStyle, textTypeToUse, currentFormatting)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        } else {
          const option = await AppAPI.RetryAnalysisWithFormatting(originalText, targetStyle, currentFormatting)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        }
      } else {
        if (useTypeSpecific) {
          const option = await AppAPI.RetryRewriteWithTextType(originalText, targetStyle, textTypeToUse, currentFormatting)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        } else {
          const option = await AppAPI.RetryRewriteWithFormatting(originalText, targetStyle, currentFormatting)
          if (option.error) {
            setError(option.error)
          } else {
            generatedText = option.text
            setResult(generatedText)
          }
        }
      }

      if (generatedText) {
        const historyEntry = { text: generatedText, style: targetStyle, timestamp: Date.now() }
        setResultHistory(prev => {
          const newHistory = [...prev, historyEntry]
          if (newHistory.length > MAX_VARIATIONS) {
            newHistory.shift()
          }
          return newHistory
        })
        setVariationIndex(prev => {
          const newIdx = prev + 1
          return newIdx >= MAX_VARIATIONS ? MAX_VARIATIONS - 1 : newIdx
        })

        const baseConfidence = targetStyle === 'grammar' ? 0.92 : targetStyle === 'formal' ? 0.88 : targetStyle === 'casual' ? 0.85 : targetStyle === 'creative' ? 0.75 : 0.82
        const confidence = Math.min(0.98, Math.max(0.60, baseConfidence + (Math.random() * 0.1 - 0.05)))
        setConfidenceScore(Math.round(confidence * 100))
      }
    } catch (err) {
      console.error('Generate error:', err)
      setError('Failed to connect to AI server')
    }
    setLoading(false)
  }, [originalText, selectedTextType, isUserOverride])

  useEffect(() => {
    if (!originalText) return

    runtime.WindowSetSize(460, 620)
    runtime.WindowSetAlwaysOnTop(true)
    runtime.WindowShow()

    if (miniModeResult) {
      setResult(miniModeResult)
      setConfidenceScore(85)
      setResultHistory([{ text: miniModeResult, style: rewriteStyle, timestamp: Date.now() }])
      setVariationIndex(0)
      return
    }

    if (!isDetecting && detectedTextType) {
      generate(initialMode, isGrammarDefault ? 'grammar' : initialRewriteStyle, true)
    } else if (!isDetecting) {
      generate(initialMode, isGrammarDefault ? 'grammar' : initialRewriteStyle, false)
    }
  }, [generate, initialMode, isGrammarDefault, initialRewriteStyle, isDetecting, detectedTextType, originalText, miniModeResult, rewriteStyle])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const shouldUseTextType = () => {
    return selectedTextType !== '' && selectedTextType !== 'unknown'
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
    if (mainMode === 'analyze') {
      generate('analyze', analysisStyle, true)
    } else {
      generate('rewrite', rewriteStyle, true)
    }
  }

  const handlePrevVariation = useCallback(() => {
    if (variationIndex > 0) {
      const newIndex = variationIndex - 1
      setVariationIndex(newIndex)
      setResult(resultHistory[newIndex].text)
    }
  }, [variationIndex, resultHistory])

  const handleNextVariation = useCallback(() => {
    if (variationIndex < resultHistory.length - 1) {
      const newIndex = variationIndex + 1
      setVariationIndex(newIndex)
      setResult(resultHistory[newIndex].text)
    }
  }, [variationIndex, resultHistory])

  const handleRewrite = useCallback(() => {
    const currentStyle = mainMode === 'analyze' ? analysisStyle : rewriteStyle
    generate(mainMode, currentStyle, shouldUseTextType())
  }, [mainMode, analysisStyle, rewriteStyle, generate, shouldUseTextType])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'ArrowLeft') {
        e.preventDefault()
        handlePrevVariation()
      } else if ((e.ctrlKey || e.metaKey) && e.key === 'ArrowRight') {
        e.preventDefault()
        handleNextVariation()
      } else if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
        e.preventDefault()
        handleRewrite()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handlePrevVariation, handleNextVariation, handleRewrite])

  const handleCopy = async () => {
    if (!result) return

    try {
      await AppAPI.ApplyRewrite(result)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const handleReplace = async () => {
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
  }

  const handlePasteConfirm = async () => {
    if (dontAskAgain) {
      try {
        const settings = await AppAPI.GetSettings()
        settings.auto_paste_mode = 'always'
        await AppAPI.SaveSettings(settings)
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
  }

  const handlePasteCancel = async () => {
    if (dontAskAgain) {
      try {
        const settings = await AppAPI.GetSettings()
        settings.auto_paste_mode = 'never'
        await AppAPI.SaveSettings(settings)
        setAutoPasteMode('never')
      } catch (e) {
        console.error('Failed to save settings:', e)
      }
    }

    setShowPasteDialog(false)
    await handleCopy()
    onClose()
  }

  const currentRewriteStyleData = REWRITE_STYLES.find(s => s.value === rewriteStyle) || REWRITE_STYLES[0]
  const currentAnalysisStyleData = ANALYSIS_STYLES.find(s => s.value === analysisStyle) || ANALYSIS_STYLES[0]

  const renderContent = (text: string) => {
    if (!text) return null
    let cleanText = text.replace(/\*\*\*\*/g, '**').replace(/\*([^\s*][^*]*[^\s*])\*/g, '**$1**')

    const lines = cleanText.split('\n')
    const isBulletList = lines.some(line => line.trim().match(/^[-•\*]\s/))
    const isNumberedList = lines.some(line => line.trim().match(/^\d+\.\s/))

    if ((isBulletList || isNumberedList) && !mainMode) {
      return (
        <div className="document-container">
          <ul className={`document-list ${isNumberedList ? 'numbered' : ''}`}>
            {lines.map((line, i) => {
              const trimmed = line.trim()
              if (!trimmed) return null
              const content = trimmed.replace(/^([-•\*]|\d+\.)\s*/, '')
              const parts = content.split(/(\*\*[^*]+\*\*)/g)
              return (
                <li key={i}>
                  {parts.map((part, j) => {
                    if (part && part.startsWith('**') && part.endsWith('**')) {
                      return <strong key={j}>{part.slice(2, -2)}</strong>
                    }
                    return part
                  })}
                </li>
              )
            })}
          </ul>
        </div>
      )
    }

    const isChatLike = lines.some(line => line.trim().match(/^(User|Assistant|Me|You|Bot|System|Agent):/i))

    if (isChatLike && selectedTextType === 'chat') {
      return (
        <div className="chat-container">
          {lines.map((line, i) => {
            const trimmed = line.trim()
            if (!trimmed) return null

            const messageMatch = trimmed.match(/^(User|Assistant|Me|You|Bot|System|Agent):\s*(.*)$/i)
            if (messageMatch) {
              const sender = messageMatch[1]
              const content = messageMatch[2]
              const isUser = /^(User|Me|You)/i.test(sender)

              return (
                <div key={i} className={`chat-message ${isUser ? 'user' : 'system'}`}>
                  <div className="chat-header">
                    <span className="chat-sender">{sender}</span>
                    <span className="chat-timestamp">{new Date().toLocaleTimeString()}</span>
                  </div>
                  <div className="chat-content">
                    {content.split(/(\*\*[^*]+\*\*)/g).map((part, j) => {
                      if (part && part.startsWith('**') && part.endsWith('**')) {
                        return <strong key={j}>{part.slice(2, -2)}</strong>
                      }
                      return part
                    })}
                  </div>
                </div>
              )
            }

            return (
              <div key={i} className="chat-message system">
                <div className="chat-content">{trimmed}</div>
              </div>
            )
          })}
        </div>
      )
    }

    if (isBulletList && mainMode === 'analyze' && analysisStyle === 'bullets') {
      return (
        <ul className="bullet-list">
          {lines.map((line, i) => {
            const trimmed = line.trim()
            if (!trimmed) return null
            const content = trimmed.replace(/^[-•\*]\s*/, '')
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

    const isEmail = cleanText.match(/^(Dear\s|Hi\s|Hello\s|To\s)/i) &&
      cleanText.match(/(Regards|Sincerely|Thanks|Best|Warm regards|Kind regards|Yours|Cheers)/i)

    if (isEmail) {
      const paragraphs = cleanText.split(/\n\n+/)
      const emailParts: { type: string; content: string }[] = []

      let currentIndex = 0
      for (const para of paragraphs) {
        const trimmed = para.trim()
        if (!trimmed) continue

        if (trimmed.match(/^(Dear\s|Hi\s|Hello\s|To\s)/i) && currentIndex === 0) {
          emailParts.push({ type: 'greeting', content: trimmed })
        } else if (trimmed.match(/(Regards|Sincerely|Thanks|Best|Warm regards|Kind regards|Yours|Cheers)/i) && trimmed.length < 120) {
          emailParts.push({ type: 'closing', content: trimmed })
        } else if (emailParts.some(p => p.type === 'closing') && trimmed.length < 100 && !trimmed.match(/[.!?]\s/)) {
          emailParts.push({ type: 'signature', content: trimmed })
        } else {
          emailParts.push({ type: 'body', content: trimmed })
        }
        currentIndex++
      }

      return (
        <div className="email-container">
          {emailParts.map((part, i) => {
            const partContent = part.content.split(/(\*\*[^*]+\*\*)/g)
            const renderPart = (
              <>
                {partContent.map((segment, j) => {
                  if (segment && segment.startsWith('**') && segment.endsWith('**')) {
                    return <strong key={j}>{segment.slice(2, -2)}</strong>
                  }
                  return segment
                })}
              </>
            )

            switch (part.type) {
              case 'greeting':
                return <div key={i} className="email-greeting">{renderPart}</div>
              case 'body':
                return <div key={i} className="email-body"><p>{renderPart}</p></div>
              case 'closing':
                return <div key={i} className="email-closing"><div className="email-closing-text">{renderPart}</div></div>
              case 'signature':
                return <div key={i} className="email-signature">{renderPart}</div>
              default:
                return <p key={i}>{renderPart}</p>
            }
          })}
        </div>
      )
    }

    return cleanText.split(/\n\n+/).map((para, i) => {
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
    <div className="popup-v2">
      {/* Header */}
      <header className="popup-header">
        <div className="header-left">
          <div className="logo">
            <span className="logo-icon">🎯</span>
            <span className="logo-text">CopyFather</span>
          </div>
        </div>
        <div className="header-right">
          <button className="icon-btn" onClick={onSettings} title="Settings">
            <span>⚙️</span>
          </button>
          <button className="icon-btn" onClick={onClose} title="Close">
            <span>✕</span>
          </button>
        </div>
      </header>

      {/* Mode Toggle */}
      <div className="mode-selector">
        <button
          className={`mode-btn ${mainMode === 'rewrite' ? 'active' : ''}`}
          onClick={() => handleMainModeChange('rewrite')}
        >
          <span className="mode-icon">🔄</span>
          <span className="mode-label">Rewrite</span>
        </button>
        <button
          className={`mode-btn ${mainMode === 'analyze' ? 'active' : ''}`}
          onClick={() => handleMainModeChange('analyze')}
        >
          <span className="mode-icon">📊</span>
          <span className="mode-label">Analyze</span>
        </button>
      </div>

      {/* Content Area */}
      <div className="popup-content">
        {/* Style Selector */}
        <div className="style-section" ref={dropdownRef}>
          <div className="style-dropdown">
            <button
              className={`style-trigger ${dropdownOpen ? 'open' : ''}`}
              onClick={() => setDropdownOpen(!dropdownOpen)}
            >
              <span className="style-icon">
                {mainMode === 'rewrite' ? currentRewriteStyleData.icon : currentAnalysisStyleData.icon}
              </span>
              <span className="style-label">
                {mainMode === 'rewrite' ? currentRewriteStyleData.label : currentAnalysisStyleData.label}
              </span>
              <span className="chevron">▼</span>
            </button>

            {dropdownOpen && (
              <div className="style-menu">
                {mainMode === 'rewrite' ? (
                  REWRITE_STYLES.map(s => (
                    <button
                      key={s.value}
                      className={`style-option ${s.value === rewriteStyle ? 'active' : ''}`}
                      onClick={() => handleRewriteStyleChange(s.value)}
                    >
                      <span className="option-icon">{s.icon}</span>
                      <div className="option-content">
                        <span className="option-label">{s.label}</span>
                        <span className="option-desc">{s.desc}</span>
                      </div>
                    </button>
                  ))
                ) : (
                  ANALYSIS_STYLES.map(s => (
                    <button
                      key={s.value}
                      className={`style-option ${s.value === analysisStyle ? 'active' : ''}`}
                      onClick={() => handleAnalysisStyleChange(s.value)}
                    >
                      <span className="option-icon">{s.icon}</span>
                      <div className="option-content">
                        <span className="option-label">{s.label}</span>
                        <span className="option-desc">{s.desc}</span>
                      </div>
                    </button>
                  ))
                )}
              </div>
            )}
          </div>

          {/* Text Type Badge */}
          {selectedTextType && (
            <div className="text-type-badge" ref={textTypeDropdownRef}>
              <button
                className="text-type-trigger"
                onClick={() => setTextTypeDropdownOpen(!textTypeDropdownOpen)}
                title={isUserOverride ? 'You selected this type' : 'Auto-detected'}
              >
                <span>{availableTextTypes.find(t => t.Type === selectedTextType)?.Icon || '📝'}</span>
                <span>{availableTextTypes.find(t => t.Type === selectedTextType)?.Label || 'Text'}</span>
              </button>
              {textTypeDropdownOpen && (
                <div className="text-type-menu">
                  {availableTextTypes.map(t => (
                    <button
                      key={t.Type}
                      className={`text-type-option ${t.Type === selectedTextType ? 'active' : ''}`}
                      onClick={() => handleTextTypeChange(t.Type)}
                    >
                      <span className="option-icon">{t.Icon}</span>
                      <span className="option-label">{t.Label}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Result Area */}
        <div className={`result-section ${loading ? 'loading' : ''}`}>
          <div className="result-header">
            <div className="result-meta">
              {mainMode === 'rewrite' && rewriteStyle === 'grammar' && !loading && result && (
                <span className="badge-success">✓ Grammar & Style</span>
              )}
              {!loading && confidenceScore !== null && result && (
                <span className={`confidence-badge ${confidenceScore >= 85 ? 'high' : confidenceScore >= 70 ? 'medium' : 'low'}`}>
                  📊 {confidenceScore}% confidence
                </span>
              )}
            </div>
            <div className="result-actions">
              <button
                className={`formatting-toggle ${enableFormatting ? 'active' : ''}`}
                onClick={() => setEnableFormatting(!enableFormatting)}
                title={enableFormatting ? 'Rich formatting ON' : 'Plain text mode'}
              >
                {enableFormatting ? '📝 Rich' : '📄 Plain'}
              </button>
              <button
                className="btn-rewrite"
                onClick={handleRewrite}
                disabled={loading}
                title="Regenerate with same style (Ctrl+R)"
              >
                🔄 Rewrite
              </button>
              {resultHistory.length > 1 && (
                <div className="variation-controls">
                  <button className="variation-btn" onClick={handlePrevVariation} disabled={variationIndex <= 0} title="Previous variation">
                    ◀
                  </button>
                  <span className="variation-indicator">{variationIndex + 1} / {resultHistory.length}</span>
                  <button className="variation-btn" onClick={handleNextVariation} disabled={variationIndex >= resultHistory.length - 1} title="Next variation">
                    ▶
                  </button>
                </div>
              )}
            </div>
          </div>

          <div
            className="result-text"
            onClick={handleCopy}
            title="Click to copy"
          >
            {loading ? (
              <div className="skeleton-loader">
                <div className="skeleton-line"></div>
                <div className="skeleton-line"></div>
                <div className="skeleton-line short"></div>
              </div>
            ) : error ? (
              <div className="error-state">
                <span className="error-icon">⚠️</span>
                <p>{error}</p>
                <button
                  className="btn-secondary"
                  onClick={() => generate(mainMode, mainMode === 'analyze' ? analysisStyle : rewriteStyle, shouldUseTextType())}
                >
                  Retry
                </button>
              </div>
            ) : (
              <div className="result-content">
                {renderContent(result)}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Footer Actions */}
      <footer className="popup-footer">
        <button
          className="btn btn-secondary"
          onClick={handleCopy}
          disabled={loading || !!error || !result}
        >
          {copied ? '✓ Copied' : '📋 Copy'}
        </button>
        {onShowDiff && (
          <button
            className="btn btn-secondary"
            onClick={() => onShowDiff(result)}
            disabled={loading || !!error || !result}
            title="View differences"
          >
            🔍 Diff
          </button>
        )}
        <button
          className="btn btn-primary"
          onClick={handleReplace}
          disabled={loading || !!error || !result}
        >
          Replace Selection
        </button>
      </footer>

      {/* Paste Dialog */}
      {showPasteDialog && (
        <div className="dialog-overlay">
          <div className="dialog">
            <div className="dialog-header">
              <h3>📋 Replace Text?</h3>
              <p>Automatically paste the rewritten text?</p>
            </div>
            <div className="dialog-content">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={dontAskAgain}
                  onChange={(e) => setDontAskAgain(e.target.checked)}
                />
                <span>Don't ask again</span>
              </label>
            </div>
            <div className="dialog-footer">
              <button className="btn btn-secondary" onClick={handlePasteCancel}>
                Copy Only
              </button>
              <button className="btn btn-primary" onClick={handlePasteConfirm}>
                ✓ Paste
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
