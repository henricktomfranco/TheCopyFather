import { useState, useCallback, useRef, useEffect, MutableRefObject } from 'react'
import * as runtime from '../../wailsjs/runtime'
import * as RewriteAPI from '../../wailsjs/go/main/RewriteService'

interface UseGenerateRewriteProps {
  originalText: string
  selectedTextType: string
  isUserOverride: boolean
  enableFormattingRef: MutableRefObject<boolean>
  mainMode: 'rewrite' | 'analyze'
  rewriteStyle: string
  analysisStyle: string
}

export function useGenerateRewrite({
  originalText,
  selectedTextType,
  isUserOverride,
  enableFormattingRef,
  mainMode,
  rewriteStyle,
  analysisStyle
}: UseGenerateRewriteProps) {
  const [result, setResult] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [confidenceScore, setConfidenceScore] = useState<number | null>(null)
  const [resultHistory, setResultHistory] = useState<Array<{ text: string; style: string; timestamp: number }>>([])
  const [variationIndex, setVariationIndex] = useState<number>(-1)
  
  const MAX_VARIATIONS = 20
  const styleCacheRef = useRef<Map<string, string>>(new Map())
  const activeRequestIDRef = useRef<string | null>(null)
  const cleanupRef = useRef<(() => void) | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
        debounceRef.current = null
      }
      if (activeRequestIDRef.current) {
        RewriteAPI.CancelStream(activeRequestIDRef.current)
        activeRequestIDRef.current = null
      }
      if (cleanupRef.current) {
        cleanupRef.current()
        cleanupRef.current = null
      }
    }
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

    if (activeRequestIDRef.current) {
      RewriteAPI.CancelStream(activeRequestIDRef.current)
      if (cleanupRef.current) {
        cleanupRef.current()
        cleanupRef.current = null
      }
      activeRequestIDRef.current = null
    }

    setLoading(true)
    setError(null)
    setResult('')

    const requestID = crypto.randomUUID()
    activeRequestIDRef.current = requestID

    const cleanup = () => {
      runtime.EventsOff(`stream:chunk:${requestID}`)
      runtime.EventsOff(`stream:done:${requestID}`)
      runtime.EventsOff(`stream:error:${requestID}`)
    }
    cleanupRef.current = cleanup

    try {
      let generatedText = ''
      const textTypeToUse = useTextType ? currentTextType : 'normal'
      const useTypeSpecific = useTextType

      await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => {
          cleanup()
          activeRequestIDRef.current = null
          reject(new Error('Streaming timeout'))
        }, 120000)

        runtime.EventsOn(`stream:chunk:${requestID}`, (chunk: string) => {
          if (chunk) {
            generatedText = chunk
            setResult(chunk)
          }
        })

        runtime.EventsOn(`stream:done:${requestID}`, () => {
          clearTimeout(timeout)
          cleanup()
          activeRequestIDRef.current = null
          resolve()
        })

        runtime.EventsOn(`stream:error:${requestID}`, (errMsg: string) => {
          clearTimeout(timeout)
          cleanup()
          activeRequestIDRef.current = null
          reject(new Error(errMsg))
        })

        if (targetMainMode === 'analyze') {
          if (useTypeSpecific) {
            RewriteAPI.StreamAnalysisWithTextType(requestID, originalText, targetStyle, textTypeToUse, currentFormatting)
          } else {
            RewriteAPI.StreamAnalysisWithTextType(requestID, originalText, targetStyle, 'normal', currentFormatting)
          }
        } else {
          if (useTypeSpecific) {
            RewriteAPI.StreamRewriteWithTextType(requestID, originalText, targetStyle, textTypeToUse, currentFormatting)
          } else {
            RewriteAPI.StreamRewriteWithFormatting(requestID, originalText, targetStyle, currentFormatting)
          }
        }
      })

      if (generatedText) {
        cache.set(cacheKey, generatedText)
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
        setConfidenceScore(Math.round(baseConfidence * 100))
      }
    } catch (err) {
      console.error('Generate error:', err)
      setError(err instanceof Error ? err.message : 'Failed to connect to AI server')
    }
    setLoading(false)
  }, [originalText, selectedTextType, isUserOverride])

  const debouncedGenerate = useCallback((mode: string, style: string, useTextType: boolean) => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    debounceRef.current = setTimeout(() => {
      generate(mode, style, useTextType)
    }, 200)
  }, [generate])

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

  const shouldUseTextType = useCallback(() => {
    return selectedTextType !== '' && selectedTextType !== 'unknown'
  }, [selectedTextType])

  const handleRewrite = useCallback(() => {
    const currentStyle = mainMode === 'analyze' ? analysisStyle : rewriteStyle
    generate(mainMode, currentStyle, shouldUseTextType())
  }, [mainMode, analysisStyle, rewriteStyle, generate, shouldUseTextType])

  return {
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
  }
}
