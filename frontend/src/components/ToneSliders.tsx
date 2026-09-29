import React from 'react'

interface ToneSlidersProps {
  formality: number
  setFormality: (val: number) => void
  length: number
  setLength: (val: number) => void
  onApply: () => void
  loading: boolean
}

export function ToneSliders({
  formality,
  setFormality,
  length,
  setLength,
  onApply,
  loading
}: ToneSlidersProps) {
  // Compute Formality Badge Info
  const getFormalityTier = (val: number) => {
    if (val <= 20) return { label: 'Ultra Casual', icon: '🏖️', desc: 'Chill, slang-friendly, conversational', color: '#38bdf8' }
    if (val <= 40) return { label: 'Friendly & Casual', icon: '💬', desc: 'Warm peer-to-peer tone', color: '#4ade80' }
    if (val <= 60) return { label: 'Balanced & Direct', icon: '⚖️', desc: 'Natural, clear, straightforward', color: '#a78bfa' }
    if (val <= 80) return { label: 'Professional', icon: '💼', desc: 'Polished corporate tone', color: '#f59e0b' }
    return { label: 'Executive & Formal', icon: '🎩', desc: 'Diplomatic, authoritative & elegant', color: '#f43f5e' }
  }

  // Compute Length Badge Info
  const getLengthTier = (val: number) => {
    if (val <= 20) return { label: 'Ultra-Short (-60%)', icon: '⚡', desc: 'Bare essence, zero fluff', color: '#38bdf8' }
    if (val <= 40) return { label: 'Concise (-30%)', icon: '✂️', desc: 'Tight sentences, punchy', color: '#4ade80' }
    if (val <= 60) return { label: 'Same Length (1:1)', icon: '📏', desc: 'Preserves original length', color: '#a78bfa' }
    if (val <= 80) return { label: 'Elaborate (+30%)', icon: '📝', desc: 'Adds helpful nuance & detail', color: '#f59e0b' }
    return { label: 'Comprehensive (+60%)', icon: '📚', desc: 'Deep dive, rich explanations', color: '#f43f5e' }
  }

  const formalityTier = getFormalityTier(formality)
  const lengthTier = getLengthTier(length)

  const PRESETS = [
    { label: '💼 Executive', formality: 85, length: 50 },
    { label: '💬 Slack / Chat', formality: 20, length: 30 },
    { label: '⚡ Punchy Pitch', formality: 60, length: 15 },
    { label: '🤝 Polite Followup', formality: 55, length: 65 },
    { label: '📚 Deep Dive', formality: 75, length: 90 },
  ]

  const applyPreset = (f: number, l: number) => {
    setFormality(f)
    setLength(l)
  }

  return (
    <div className="tone-sliders-container">
      {/* Dual Sliders Responsive Grid */}
      <div className="sliders-cards-grid">
        {/* Formality Slider Card */}
        <div className="slider-card">
          <div className="slider-card-header">
            <div className="slider-label-group">
              <span className="slider-title">Formality</span>
              <span className="slider-badge" style={{ backgroundColor: `${formalityTier.color}22`, color: formalityTier.color, borderColor: `${formalityTier.color}44` }}>
                {formalityTier.icon} {formalityTier.label}
              </span>
            </div>
            <span className="slider-value">{formality}%</span>
          </div>

          <div className="slider-track-wrap">
            <input
              type="range"
              min="0"
              max="100"
              step="5"
              value={formality}
              onChange={(e) => setFormality(parseInt(e.target.value, 10))}
              className="modern-range-slider"
              style={{
                background: `linear-gradient(90deg, #10b981 0%, ${formalityTier.color} ${formality}%, rgba(255,255,255,0.1) ${formality}%)`
              }}
            />
          </div>

          <div className="slider-endpoints">
            <span>Casual</span>
            <span>Balanced</span>
            <span>Formal</span>
          </div>
        </div>

        {/* Length Slider Card */}
        <div className="slider-card">
          <div className="slider-card-header">
            <div className="slider-label-group">
              <span className="slider-title">Length</span>
              <span className="slider-badge" style={{ backgroundColor: `${lengthTier.color}22`, color: lengthTier.color, borderColor: `${lengthTier.color}44` }}>
                {lengthTier.icon} {lengthTier.label}
              </span>
            </div>
            <span className="slider-value">{length}%</span>
          </div>

          <div className="slider-track-wrap">
            <input
              type="range"
              min="0"
              max="100"
              step="5"
              value={length}
              onChange={(e) => setLength(parseInt(e.target.value, 10))}
              className="modern-range-slider length-slider"
              style={{
                background: `linear-gradient(90deg, #6366f1 0%, ${lengthTier.color} ${length}%, rgba(255,255,255,0.1) ${length}%)`
              }}
            />
          </div>

          <div className="slider-endpoints">
            <span>Concise</span>
            <span>Original</span>
            <span>Expanded</span>
          </div>
        </div>
      </div>

      {/* Quick Presets Pills */}
      <div className="presets-row">
        <span className="presets-caption">Quick Jumps:</span>
        <div className="presets-list">
          {PRESETS.map((p) => {
            const isActive = formality === p.formality && length === p.length
            return (
              <button
                key={p.label}
                type="button"
                className={`preset-chip ${isActive ? 'active' : ''}`}
                onClick={() => applyPreset(p.formality, p.length)}
              >
                {p.label}
              </button>
            )
          })}
        </div>
      </div>

      {/* Apply Button */}
      <button
        type="button"
        className="btn-apply-sliders"
        onClick={onApply}
        disabled={loading}
      >
        {loading ? (
          <>
            <span className="spinner-small" /> Generating with Custom Tone...
          </>
        ) : (
          <>
            ✨ Apply Tone & Length
          </>
        )}
      </button>
    </div>
  )
}
