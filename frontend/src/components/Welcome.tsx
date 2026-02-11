import { useState, useEffect } from 'react'
import * as runtime from '../../wailsjs/runtime'
import '../styles/Popup.css' // Reuse premium styles

interface WelcomeProps {
    onAccept: () => void
}

function Welcome({ onAccept }: WelcomeProps) {
    useEffect(() => {
        runtime.WindowSetSize(500, 520)
        runtime.WindowCenter()
        runtime.WindowShow()
    }, [])

    return (
        <div className="popup modern welcome-screen">
            <div className="welcome-header">
                <div className="welcome-icon">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M12 2L2 7l10 5 10-5-10-5z" stroke="url(#gradient1)"/>
                        <path d="M2 17l10 5 10-5" stroke="url(#gradient1)"/>
                        <path d="M2 12l10 5 10-5" stroke="url(#gradient1)"/>
                        <defs>
                            <linearGradient id="gradient1" x1="0%" y1="0%" x2="100%" y2="100%">
                                <stop offset="0%" stopColor="#6366f1" />
                                <stop offset="100%" stopColor="#8b5cf6" />
                            </linearGradient>
                        </defs>
                    </svg>
                </div>
                <h2 className="gradient-text">Welcome to The Copy Father</h2>
                <p className="welcome-subtitle">Network Access Required</p>
            </div>

            <div className="welcome-content">
                <div className="welcome-card">
                    <p className="welcome-description">
                        Connect to an <strong>Ollama node</strong> to unlock AI-powered text rewriting
                    </p>
                    <p className="welcome-note">
                        This involves local or remote network access. By continuing, you agree to allow communication with your Ollama server.
                    </p>
                </div>

                <div className="privacy-notice">
                    <div className="privacy-icon">🔒</div>
                    <div className="privacy-text">
                        <strong>Privacy Notice:</strong> Your text is sent to your configured Ollama server. It's not stored by this app, but ensure you trust your server location.
                    </div>
                </div>

                <div className="permissions-list">
                    <div className="permission-item allowed">
                        <span className="permission-icon">✓</span>
                        <span>Communicate with Ollama API</span>
                    </div>
                    <div className="permission-item allowed">
                        <span className="permission-icon">✓</span>
                        <span>Fetch available AI models</span>
                    </div>
                    <div className="permission-item warning">
                        <span className="permission-icon">⚠</span>
                        <span>Send text to Ollama server</span>
                    </div>
                </div>
            </div>

            <div className="modern-footer">
                <button
                    className="primary-btn"
                    onClick={onAccept}
                >
                    Grant Access & Continue
                </button>
            </div>

            <style>{`
                .welcome-screen {
                    justify-content: space-between;
                }
                
                .welcome-header {
                    text-align: center;
                    padding-bottom: 24px;
                    border-bottom: 1px solid var(--border-subtle);
                }
                
                .welcome-icon {
                    width: 64px;
                    height: 64px;
                    margin: 0 auto 20px;
                    background: var(--bg-glass);
                    backdrop-filter: var(--glass-backdrop);
                    border: 1px solid var(--border-subtle);
                    border-radius: var(--radius-lg);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    animation: float 3s ease-in-out infinite;
                }
                
                .welcome-header h2 {
                    font-family: var(--font-display);
                    font-size: 24px;
                    font-weight: 600;
                    margin: 0 0 8px 0;
                }
                
                .welcome-subtitle {
                    color: var(--text-secondary);
                    font-size: 14px;
                    margin: 0;
                }
                
                .welcome-content {
                    flex: 1;
                    display: flex;
                    flex-direction: column;
                    gap: 20px;
                    padding: 24px 0;
                }
                
                .welcome-card {
                    background: var(--bg-glass);
                    backdrop-filter: var(--glass-backdrop);
                    border: 1px solid var(--border-subtle);
                    border-radius: var(--radius-md);
                    padding: 20px;
                    text-align: center;
                }
                
                .welcome-description {
                    font-size: 15px;
                    line-height: 1.6;
                    color: var(--text-primary);
                    margin: 0 0 12px 0;
                }
                
                .welcome-description strong {
                    color: var(--text-accent);
                    font-weight: 600;
                }
                
                .welcome-note {
                    font-size: 13px;
                    line-height: 1.5;
                    color: var(--text-muted);
                    margin: 0;
                }
                
                .privacy-notice {
                    display: flex;
                    gap: 12px;
                    background: rgba(99, 102, 241, 0.1);
                    border: 1px solid rgba(99, 102, 241, 0.2);
                    border-radius: var(--radius-md);
                    padding: 16px;
                    backdrop-filter: blur(10px);
                }
                
                .privacy-icon {
                    font-size: 20px;
                    flex-shrink: 0;
                }
                
                .privacy-text {
                    font-size: 12px;
                    line-height: 1.5;
                    color: var(--text-secondary);
                }
                
                .privacy-text strong {
                    color: var(--text-accent);
                }
                
                .permissions-list {
                    display: flex;
                    flex-direction: column;
                    gap: 10px;
                }
                
                .permission-item {
                    display: flex;
                    gap: 12px;
                    align-items: center;
                    padding: 12px 16px;
                    background: var(--bg-glass);
                    border: 1px solid var(--border-subtle);
                    border-radius: var(--radius-sm);
                    font-size: 13px;
                    color: var(--text-secondary);
                    transition: var(--transition-smooth);
                }
                
                .permission-item:hover {
                    background: var(--bg-glass-hover);
                }
                
                .permission-icon {
                    font-size: 14px;
                    font-weight: bold;
                    width: 20px;
                    text-align: center;
                }
                
                .permission-item.allowed .permission-icon {
                    color: #4ade80;
                }
                
                .permission-item.warning .permission-icon {
                    color: #fbbf24;
                }
                
                @keyframes float {
                    0%, 100% { transform: translateY(0); }
                    50% { transform: translateY(-8px); }
                }
            `}</style>
        </div>
    )
}

export default Welcome
