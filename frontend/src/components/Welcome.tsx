import { useState, useEffect } from 'react'
import * as runtime from '../../wailsjs/runtime'
import * as SettingsAPI from '../../wailsjs/go/main/SettingsService'
import appIcon from '../assets/appicon.png'
import '../styles/Popup.css'

interface WelcomeProps {
    onAccept: () => void
}

interface SetupStatus {
    engine_installed: boolean
    engine_path: string
    model_installed: boolean
    model_path: string
    is_downloading: boolean
    current_step: string
    progress_percent: number
    bytes_downloaded: number
    total_bytes: number
    speed: string
    error?: string
}

function Welcome({ onAccept }: WelcomeProps) {
    const [setupStatus, setSetupStatus] = useState<SetupStatus | null>(null)
    const [isDownloading, setIsDownloading] = useState(false)
    const [downloadProgress, setDownloadProgress] = useState(0)
    const [downloadSpeed, setDownloadSpeed] = useState('')
    const [currentStep, setCurrentStep] = useState<string>('idle')
    const [bytesInfo, setBytesInfo] = useState<string>('')
    const [errorMessage, setErrorMessage] = useState<string>('')
    const [isFinished, setIsFinished] = useState(false)

    useEffect(() => {
        runtime.WindowSetSize(520, 580)
        runtime.WindowCenter()
        runtime.WindowShow()

        // Check initial status
        refreshStatus()

        // Listen for download events from Go backend
        runtime.EventsOn('setup:status', (status: SetupStatus) => {
            setSetupStatus(status)
            setIsDownloading(status.is_downloading)
            setDownloadProgress(Math.round(status.progress_percent || 0))
            setDownloadSpeed(status.speed || '')
            setCurrentStep(status.current_step || 'downloading')

            if (status.total_bytes > 0) {
                const dlMB = (status.bytes_downloaded / (1024 * 1024)).toFixed(1)
                const totalMB = (status.total_bytes / (1024 * 1024)).toFixed(1)
                setBytesInfo(`${dlMB} MB / ${totalMB} MB`)
            }
        })

        runtime.EventsOn('setup:completed', () => {
            setIsDownloading(false)
            setIsFinished(true)
            setCurrentStep('completed')
            refreshStatus()
        })

        runtime.EventsOn('setup:error', (err: string) => {
            setIsDownloading(false)
            setErrorMessage(err || 'Failed downloading files.')
        })

        return () => {
            runtime.EventsOff('setup:status')
            runtime.EventsOff('setup:completed')
            runtime.EventsOff('setup:error')
        }
    }, [])

    const refreshStatus = async () => {
        try {
            const status = await SettingsAPI.CheckSetupStatus()
            setSetupStatus(status)
            if (status.engine_installed && status.model_installed) {
                setIsFinished(true)
            } else if (!status.is_downloading) {
                // Auto-start download immediately when opening on a fresh PC
                handleStartDownload()
            }
        } catch (e) {
            console.error('Failed to check setup status:', e)
        }
    }

    const handleStartDownload = async () => {
        setErrorMessage('')
        setIsDownloading(true)
        setIsFinished(false)
        try {
            await SettingsAPI.StartAutoDownload()
        } catch (err: any) {
            setIsDownloading(false)
            setErrorMessage(err?.message || String(err))
        }
    }

    const handleCancel = () => {
        SettingsAPI.CancelAutoDownload()
        setIsDownloading(false)
        setCurrentStep('idle')
    }

    const bothReady = (setupStatus?.engine_installed && setupStatus?.model_installed) || isFinished

    const getStepLabel = () => {
        switch (currentStep) {
            case 'engine':
                return 'Downloading embedded llama.cpp engine (~18 MB)...'
            case 'extracting':
                return 'Extracting engine binaries & CPU kernels...'
            case 'model':
                return 'Downloading SmolLM2-360M Q4_K_M GGUF model (~258 MB)...'
            case 'completed':
                return '✓ All files installed & engine ready!'
            default:
                return 'Preparing download...'
        }
    }

    return (
        <div className="popup modern welcome-screen">
            <div className="welcome-header">
                <div className="welcome-icon">
                    <img src={appIcon} alt="CopyFather" style={{ width: 64, height: 64 }} />
                </div>
                <h1>Welcome to The Copyfather</h1>
                <p className="welcome-subtitle">Self-Contained Local AI Assistant</p>
            </div>

            <div className="welcome-content" style={{ padding: '0 24px' }}>
                <p style={{ fontSize: '13px', color: '#cbd5e1', marginBottom: '16px', lineHeight: '1.5' }}>
                    TheCopyFather runs <strong>100% locally and privately</strong> inside this application without requiring Ollama, Docker, or any external software.
                </p>

                {/* Status Cards */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginBottom: '18px' }}>
                    {/* Engine Card */}
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '12px 16px',
                        background: 'rgba(255, 255, 255, 0.05)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '8px'
                    }}>
                        <div>
                            <strong style={{ fontSize: '13px', display: 'block' }}>💻 Embedded llama.cpp Engine</strong>
                            <span style={{ fontSize: '11px', color: '#94a3b8' }}>Windows CPU inference (~18 MB)</span>
                        </div>
                        <div>
                            {setupStatus?.engine_installed || isFinished ? (
                                <span style={{ color: '#48bb78', fontWeight: 600, fontSize: '12px' }}>✓ Ready</span>
                            ) : (
                                <span style={{ color: '#f6ad55', fontWeight: 600, fontSize: '12px' }}>⬇ Missing</span>
                            )}
                        </div>
                    </div>

                    {/* Model Card */}
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '12px 16px',
                        background: 'rgba(255, 255, 255, 0.05)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '8px'
                    }}>
                        <div>
                            <strong style={{ fontSize: '13px', display: 'block' }}>🧠 SmolLM2-360M Model</strong>
                            <span style={{ fontSize: '11px', color: '#94a3b8' }}>Q4_K_M GGUF (~258 MB, ~200 MB RAM)</span>
                        </div>
                        <div>
                            {setupStatus?.model_installed || isFinished ? (
                                <span style={{ color: '#48bb78', fontWeight: 600, fontSize: '12px' }}>✓ Ready</span>
                            ) : (
                                <span style={{ color: '#f6ad55', fontWeight: 600, fontSize: '12px' }}>⬇ Missing</span>
                            )}
                        </div>
                    </div>
                </div>

                {/* Progress Section */}
                {isDownloading && (
                    <div style={{
                        padding: '14px',
                        background: 'rgba(59, 130, 246, 0.1)',
                        border: '1px solid rgba(59, 130, 246, 0.25)',
                        borderRadius: '8px',
                        marginBottom: '16px'
                    }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '8px' }}>
                            <span style={{ fontWeight: 600 }}>{getStepLabel()}</span>
                            <span>{downloadProgress}%</span>
                        </div>

                        {/* Progress Bar */}
                        <div style={{
                            width: '100%',
                            height: '8px',
                            background: 'rgba(255, 255, 255, 0.15)',
                            borderRadius: '4px',
                            overflow: 'hidden',
                            marginBottom: '6px'
                        }}>
                            <div style={{
                                width: `${downloadProgress}%`,
                                height: '100%',
                                background: 'linear-gradient(90deg, #3b82f6, #60a5fa)',
                                transition: 'width 0.2s ease',
                                borderRadius: '4px'
                            }} />
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: '#94a3b8' }}>
                            <span>{bytesInfo}</span>
                            <span>{downloadSpeed}</span>
                        </div>
                    </div>
                )}

                {/* Error Banner */}
                {errorMessage && (
                    <div style={{
                        padding: '10px 14px',
                        background: 'rgba(239, 68, 68, 0.15)',
                        border: '1px solid rgba(239, 68, 68, 0.3)',
                        borderRadius: '6px',
                        color: '#fca5a5',
                        fontSize: '12px',
                        marginBottom: '14px'
                    }}>
                        ⚠️ {errorMessage}
                    </div>
                )}
            </div>

            <div className="modern-footer" style={{ padding: '16px 24px', display: 'flex', gap: '10px' }}>
                {bothReady ? (
                    <button
                        className="primary-btn"
                        style={{ width: '100%', padding: '12px' }}
                        onClick={onAccept}
                    >
                        Launch TheCopyFather →
                    </button>
                ) : isDownloading ? (
                    <button
                        className="btn btn-secondary"
                        style={{ width: '100%', padding: '10px' }}
                        onClick={handleCancel}
                    >
                        Cancel Download
                    </button>
                ) : (
                    <button
                        className="primary-btn"
                        style={{ width: '100%', padding: '12px' }}
                        onClick={handleStartDownload}
                    >
                        Download Files Automatically (~1.1 GB)
                    </button>
                )}
            </div>

            <style>{`
                .welcome-screen {
                    justify-content: space-between;
                }
                .welcome-header {
                    text-align: center;
                    padding: 24px 20px 16px;
                }
                .welcome-icon {
                    width: 64px;
                    height: 64px;
                    margin: 0 auto 12px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                }
                .welcome-header h1 {
                    font-size: 20px;
                    font-weight: 600;
                    margin: 0 0 4px 0;
                    color: var(--text-primary);
                }
                .welcome-subtitle {
                    font-size: 13px;
                    color: #94a3b8;
                    margin: 0;
                }
            `}</style>
        </div>
    )
}

export default Welcome
