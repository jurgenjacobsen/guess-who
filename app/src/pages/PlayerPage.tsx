import { useState, useEffect } from 'react'

interface Person {
  id: string
  name: string
  photo: string
}

export default function PlayerPage() {
  const [people, setPeople] = useState<Person[]>([])
  const [folded, setFolded] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/people')
      .then(res => res.json())
      .then(data => {
        setPeople(data ?? [])
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  function toggleFold(id: string) {
    setFolded(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  if (loading) return <div className="state-msg">Loading…</div>
  if (people.length === 0) return <div className="state-msg">No people added yet. Ask the moderator to add some!</div>

  return (
    <div className="player-grid">
      {people.map(person => (
        <div
          key={person.id}
          className={`card${folded.has(person.id) ? ' folded' : ''}`}
          onClick={() => toggleFold(person.id)}
          role="button"
          aria-pressed={folded.has(person.id)}
          aria-label={folded.has(person.id) ? `${person.name} — hidden` : person.name}
        >
          <div className="card-inner">
            <div className="card-front">
              <img src={person.photo} alt={person.name} />
              <span className="card-name">{person.name}</span>
            </div>
            <div className="card-back">
              <svg viewBox="0 0 60 80" aria-hidden="true">
                <rect width="60" height="80" rx="6" fill="currentColor" opacity="0.12" />
                <text x="50%" y="55%" dominantBaseline="middle" textAnchor="middle" fontSize="36" fill="currentColor">?</text>
              </svg>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
