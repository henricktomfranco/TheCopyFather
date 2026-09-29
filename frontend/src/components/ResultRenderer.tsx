interface ResultRendererProps {
  result: string
  mainMode: 'rewrite' | 'sliders' | 'analyze'
  analysisStyle: string
  selectedTextType: string
}

export function ResultRenderer({ result, mainMode, analysisStyle, selectedTextType }: ResultRendererProps) {
  if (!result) return null

  let cleanText = result.replace(/\*\*\*\*/g, '**').replace(/\*([^\s*][^*]*[^\s*])\*/g, '**$1**')

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
               return <p key={i} className="email-body">{renderPart}</p>
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

  return (
    <>
      {cleanText.split(/\n\n+/).map((para, i) => {
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
      })}
    </>
  )
}
