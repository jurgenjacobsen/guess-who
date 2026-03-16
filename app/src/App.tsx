import { useState } from 'react'
import PlayerPage from './pages/PlayerPage'
import ModeratorPage from './pages/ModeratorPage'
import './App.css'

type Page = 'player' | 'moderator'

export default function App() {
  const [page, setPage] = useState<Page>('player')

  return (
    <div className="app">
      <nav className="nav">
        <span className="nav-brand">Guess Who</span>
        <div className="nav-links">
          <button
            className={`nav-btn${page === 'player' ? ' active' : ''}`}
            onClick={() => setPage('player')}
          >
            Player
          </button>
          <button
            className={`nav-btn${page === 'moderator' ? ' active' : ''}`}
            onClick={() => setPage('moderator')}
          >
            Moderator
          </button>
        </div>
      </nav>
      <main className="main">
        {page === 'player' ? <PlayerPage /> : <ModeratorPage />}
      </main>
    </div>
  )
}


