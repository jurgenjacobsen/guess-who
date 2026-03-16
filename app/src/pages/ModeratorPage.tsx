import { useState, useEffect, useRef } from 'react'

interface Person {
  id: string
  name: string
  photo: string
}

export default function ModeratorPage() {
  const [people, setPeople] = useState<Person[]>([])
  const [loading, setLoading] = useState(true)
  const [name, setName] = useState('')
  const [photoMode, setPhotoMode] = useState<'url' | 'upload'>('url')
  const [photoUrl, setPhotoUrl] = useState('')
  const [photoData, setPhotoData] = useState('')
  const [photoPreview, setPhotoPreview] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  function loadPeople() {
    fetch('/api/people')
      .then(res => res.json())
      .then(data => {
        setPeople(data ?? [])
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }

  useEffect(() => { loadPeople() }, [])

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) {
      setPhotoData('')
      setPhotoPreview('')
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      setPhotoData(result)
      setPhotoPreview(result)
    }
    reader.readAsDataURL(file)
  }

  function switchMode(mode: 'url' | 'upload') {
    setPhotoMode(mode)
    setPhotoData('')
    setPhotoUrl('')
    setPhotoPreview('')
    if (fileRef.current) fileRef.current.value = ''
  }

  async function handleDelete(id: string) {
    try {
      await fetch('/api/people', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id }),
      })
      loadPeople()
    } catch {
      // non-critical, refresh anyway
      loadPeople()
    }
  }

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    const photo = photoMode === 'upload' ? photoData : photoUrl.trim()
    if (!name.trim()) {
      setError('Name is required.')
      return
    }
    if (!photo) {
      setError('A photo URL or uploaded file is required.')
      return
    }

    setSubmitting(true)
    try {
      const res = await fetch('/api/people', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), photo }),
      })
      if (res.ok) {
        setName('')
        setPhotoUrl('')
        setPhotoData('')
        setPhotoPreview('')
        if (fileRef.current) fileRef.current.value = ''
        loadPeople()
      } else {
        const msg = await res.text()
        setError(msg || 'Failed to add person.')
      }
    } catch {
      setError('Network error. Is the server running?')
    } finally {
      setSubmitting(false)
    }
  }

  const currentPreview = photoMode === 'url' ? photoUrl : photoPreview

  return (
    <div className="moderator-page">
      {/* ── ADD FORM ── */}
      <section className="mod-section add-section">
        <h2>Add Person</h2>
        <form className="add-form" onSubmit={handleAdd} noValidate>
          <label className="field-label">Name</label>
          <input
            type="text"
            className="field-input"
            placeholder="Full name"
            value={name}
            onChange={e => setName(e.target.value)}
            autoComplete="off"
          />

          <label className="field-label">Photo</label>
          <div className="toggle-group">
            <button
              type="button"
              className={`toggle-btn${photoMode === 'url' ? ' active' : ''}`}
              onClick={() => switchMode('url')}
            >
              URL
            </button>
            <button
              type="button"
              className={`toggle-btn${photoMode === 'upload' ? ' active' : ''}`}
              onClick={() => switchMode('upload')}
            >
              Upload
            </button>
          </div>

          {photoMode === 'url' ? (
            <input
              type="url"
              className="field-input"
              placeholder="https://example.com/photo.jpg"
              value={photoUrl}
              onChange={e => setPhotoUrl(e.target.value)}
            />
          ) : (
            <div className="file-input-wrap">
              <input
                ref={fileRef}
                type="file"
                accept="image/*"
                onChange={handleFileChange}
              />
              <p className="file-hint">Accepted: JPG, PNG, GIF, WebP — converted to data string</p>
            </div>
          )}

          {currentPreview && (
            <div className="photo-preview">
              <img src={currentPreview} alt="Preview" />
            </div>
          )}

          {error && <p className="form-error" role="alert">{error}</p>}

          <button type="submit" className="submit-btn" disabled={submitting}>
            {submitting ? 'Adding…' : 'Add Person'}
          </button>
        </form>
      </section>

      {/* ── PEOPLE LIST ── */}
      <section className="mod-section list-section">
        <h2>People ({people.length})</h2>
        {loading ? (
          <div className="state-msg">Loading…</div>
        ) : people.length === 0 ? (
          <div className="state-msg">No people added yet.</div>
        ) : (
          <ul className="people-list">
            {people.map(person => (
              <li key={person.id} className="person-item">
                <img
                  src={person.photo}
                  alt={person.name}
                  className="person-thumb"
                />
                <span className="person-name">{person.name}</span>
                <button
                  className="remove-btn"
                  onClick={() => handleDelete(person.id)}
                  aria-label={`Remove ${person.name}`}
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
