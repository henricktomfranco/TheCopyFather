import { useState, useEffect, useRef } from 'react'
import * as runtime from '../../wailsjs/runtime'
import * as SettingsAPI from '../../wailsjs/go/main/SettingsService'
import * as RewriteAPI from '../../wailsjs/go/main/RewriteService'
import { rewriter as rewriterModels } from '../../wailsjs/go/models'
import '../styles/main.css'
import '../styles/Popup.css'

import { useGenerateRewrite } from '../hooks/useGenerateRewrite'
import { useClipboardPaste } from '../hooks/useClipboardPaste'
import { PopupHeader } from './PopupHeader'
import { StyleSelector } from './StyleSelector'
import { ResultRenderer } from './ResultRenderer'
import { ActionFooter } from './ActionFooter'

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
  const [autoPasteMode, setAutoPasteMode] = useState<string>('ask')
  
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [enableFormatting, setEnableFormatting] = useState(true)
  const enableFormattingRef = useRef(enableFormatting)
  
  const [detectedTextType, setDetectedTextType] = useState<DetectedTextType | null>(null)
  const [selectedTextType, setSelectedTextType] = useState<string>('')
  const [availableTextTypes, setAvailableTextTypes] = useState<rewriterModels.TextTypeInfo[]>([])
  const [textTypeDropdownOpen, setTextTypeDropdownOpen] = useState(false)
  const [isDetecting, setIsDetecting] = useState(!!originalText)
  const [isUserOverride, setIsUserOverride] = useState(false)
  const hasInitialGeneratedRef = useRef(false)

  const dropdownRef = useRef<HTMLDivElement>(null)
  const textTypeDropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    enableFormattingRef.current = enableFormatting
  }, [enableFormatting])

  const {
    result,
    setResult,
    loading,
    error,
    confidenceScore,
    setConfidenceScore,
    resultHistory,
    setResultHistory,
    variationIndex,
    setVariationIndex,
    generate,
    debouncedGenerate,
    handlePrevVariation,
    handleNextVariation,
    handleRewrite,
    shouldUseTextType
  } = useGenerateRewrite({
    originalText,
    selectedTextType,
    isUserOverride,
    enableFormattingRef,
    mainMode,
    rewriteStyle,
    analysisStyle
  })

  const {
    copied,
    showPasteDialog,
    setShowPasteDialog,
    dontAskAgain,
    setDontAskAgain,
    handleCopy,
    handleReplace,
    handlePasteConfirm,
    handlePasteCancel
  } = useClipboardPaste({
    result,
    onClose,
    autoPasteMode,
    setAutoPasteMode
  })

  useEffect(() => {
    if (onResultChange) {
      onResultChange(result)
    }
  }, [result, onResultChange])

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const settings = await SettingsAPI.GetSettings()
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
        const types = await RewriteAPI.GetTextTypes()
        setAvailableTextTypes(types)

        if (originalText) {
          setIsDetecting(true)
          const detected = await RewriteAPI.DetectTextType(originalText)
          setDetectedTextType(detected)
          if (!isUserOverride) {
            setSelectedTextType(detected.type)
          }
          setIsDetecting(false)
        }
      } catch (e) {
        console.error('Failed to load text types:', e)
        setIsDetecting(false)
      }
    }
    loadTextTypesAndDetect()
  }, [originalText, isUserOverride])

  useEffect(() => {
    const unlisten = runtime.EventsOn('context:window', (title: string) => {
      console.log('Detected window context:', title)
      const lowerTitle = title.toLowerCase()
      let mappedType = ''
      let mappedStyle = ''

      if (lowerTitle.includes('discord') || lowerTitle.includes('slack') || lowerTitle.includes('teams') || lowerTitle.includes('messenger') || lowerTitle.includes('whatsapp')) {
        mappedType = 'chat'
        mappedStyle = 'casual'
      } else if (lowerTitle.includes('outlook') || lowerTitle.includes('mail') || lowerTitle.includes('gmail') || lowerTitle.includes('thunderbird')) {
        mappedType = 'email'
        mappedStyle = 'formal'
      } else if (lowerTitle.includes('word') || lowerTitle.includes('notepad') || lowerTitle.includes('code') || lowerTitle.includes('obsidian')) {
        mappedType = 'normal'
        mappedStyle = 'standard'
      }

      if (mappedType && mappedStyle) {
        setRewriteStyle(mappedStyle)
        setSelectedTextType(mappedType)
        setIsUserOverride(false)
      }
    })
    return () => unlisten()
  }, [])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false)
      }
      if (textTypeDropdownRef.current && !textTypeDropdownRef.current.contains(event.target as Node)) {
        setTextTypeDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  useEffect(() => {
    if (!originalText) return

    runtime.WindowSetSize(500, 680)
    runtime.WindowSetAlwaysOnTop(true)
    runtime.WindowShow()

    if (miniModeResult) {
      setResult(miniModeResult)
      setConfidenceScore(85)
      setResultHistory([{ text: miniModeResult, style: rewriteStyle, timestamp: Date.now() }])
      setVariationIndex(0)
      return
    }

    if (!isDetecting && !hasInitialGeneratedRef.current) {
      hasInitialGeneratedRef.current = true
      const useTextType = !!detectedTextType || selectedTextType !== ''
      const styleToUse = isGrammarDefault ? 'grammar' : initialRewriteStyle
      generate(initialMode, styleToUse, useTextType)
    }
  }, [generate, initialMode, isGrammarDefault, initialRewriteStyle, isDetecting, detectedTextType, selectedTextType, originalText, miniModeResult, rewriteStyle])

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
      } else if (e.key === 'Escape') {
        e.preventDefault()
        onClose()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handlePrevVariation, handleNextVariation, handleRewrite])

  const handleMainModeChange = (newMode: 'rewrite' | 'analyze') => {
    if (newMode === mainMode) return
    setMainMode(newMode)
    const useTextType = shouldUseTextType()
    if (newMode === 'analyze') {
      debouncedGenerate('analyze', analysisStyle, useTextType)
    } else {
      debouncedGenerate('rewrite', rewriteStyle, useTextType)
    }
  }

  const handleRewriteStyleChange = (newStyle: string) => {
    setRewriteStyle(newStyle)
    setDropdownOpen(false)
    debouncedGenerate('rewrite', newStyle, shouldUseTextType())
  }

  const handleAnalysisStyleChange = (newStyle: string) => {
    setAnalysisStyle(newStyle)
    setDropdownOpen(false)
    debouncedGenerate('analyze', newStyle, shouldUseTextType())
  }

  const handleTextTypeChange = (newType: string) => {
    setSelectedTextType(newType)
    setTextTypeDropdownOpen(false)
    setIsUserOverride(true)
    if (mainMode === 'analyze') {
      debouncedGenerate('analyze', analysisStyle, true)
    } else {
      debouncedGenerate('rewrite', rewriteStyle, true)
    }
  }

  const currentRewriteStyleData = REWRITE_STYLES.find(s => s.value === rewriteStyle) || REWRITE_STYLES[0]
  const currentAnalysisStyleData = ANALYSIS_STYLES.find(s => s.value === analysisStyle) || ANALYSIS_STYLES[0]

  return (
    <div className="popup-v2">
      <PopupHeader 
        onSettings={onSettings}
        onClose={onClose}
        mainMode={mainMode}
        handleMainModeChange={handleMainModeChange}
      />

      <div className="popup-content">
        <StyleSelector
          dropdownRef={dropdownRef}
          textTypeDropdownRef={textTypeDropdownRef}
          dropdownOpen={dropdownOpen}
          setDropdownOpen={setDropdownOpen}
          textTypeDropdownOpen={textTypeDropdownOpen}
          setTextTypeDropdownOpen={setTextTypeDropdownOpen}
          mainMode={mainMode}
          rewriteStyle={rewriteStyle}
          analysisStyle={analysisStyle}
          REWRITE_STYLES={REWRITE_STYLES}
          ANALYSIS_STYLES={ANALYSIS_STYLES}
          currentRewriteStyleData={currentRewriteStyleData}
          currentAnalysisStyleData={currentAnalysisStyleData}
          handleRewriteStyleChange={handleRewriteStyleChange}
          handleAnalysisStyleChange={handleAnalysisStyleChange}
          selectedTextType={selectedTextType}
          availableTextTypes={availableTextTypes}
          isUserOverride={isUserOverride}
          handleTextTypeChange={handleTextTypeChange}
        />

        <div className={`result-section ${loading ? 'loading' : ''}`}>
          <div className="result-header">
            <div className="result-meta">
              {mainMode === 'rewrite' && rewriteStyle === 'grammar' && !loading && result && (
                <span className="badge-success">✓ Grammar & Style</span>
              )}
              {loading && result && (
                <span className="streaming-badge">
                  <span className="streaming-dot"></span> Streaming...
                </span>
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
            {error ? (
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
            ) : result ? (
              <div className="result-content">
                <ResultRenderer 
                  result={result}
                  mainMode={mainMode}
                  analysisStyle={analysisStyle}
                  selectedTextType={selectedTextType}
                />
              </div>
            ) : loading ? (
              <div className="skeleton-loader">
                <div className="skeleton-line"></div>
                <div className="skeleton-line medium"></div>
                <div className="skeleton-line"></div>
                <div className="skeleton-line short"></div>
              </div>
            ) : (
              <div className="empty-state">
                <div className="empty-icon">
                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </div>
                <span>Select a style and click Rewrite</span>
              </div>
            )}
          </div>
        </div>
      </div>

      <ActionFooter 
        loading={loading}
        error={error}
        result={result}
        copied={copied}
        handleCopy={handleCopy}
        onShowDiff={onShowDiff}
        handleReplace={handleReplace}
        showPasteDialog={showPasteDialog}
        dontAskAgain={dontAskAgain}
        setDontAskAgain={setDontAskAgain}
        handlePasteCancel={handlePasteCancel}
        handlePasteConfirm={handlePasteConfirm}
      />
    </div>
  )
}
