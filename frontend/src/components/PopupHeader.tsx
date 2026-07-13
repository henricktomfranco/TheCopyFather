import appIcon from '../assets/appicon.png'

interface PopupHeaderProps {
  onSettings: () => void
  onClose: () => void
  mainMode: 'rewrite' | 'analyze'
  handleMainModeChange: (mode: 'rewrite' | 'analyze') => void
}

export function PopupHeader({ onSettings, onClose, mainMode, handleMainModeChange }: PopupHeaderProps) {
  return (
    <>
      <header className="popup-header">
        <div className="header-left">
          <div className="logo">
            <img src={appIcon} alt="CopyFather" className="logo-icon" />
            <span className="logo-text">CopyFather</span>
          </div>
        </div>
        <div className="header-right">
          <button className="icon-btn" onClick={onSettings} title="Settings">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
          </button>
          <button className="icon-btn" onClick={onClose} title="Close">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
      </header>

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
    </>
  )
}
