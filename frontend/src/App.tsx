import { FormEvent, useCallback, useEffect, useState } from 'react'

type Quote = {
  id: number
  provider: string
  service: string
  price: number
  currency: string
  pickup: string
  dropoff: string
  observedAt: string
  source: string
}

type QuoteResponse = { quotes: Quote[] }

function ageLabel(observedAt: string) {
  const minutes = Math.max(0, Math.floor((Date.now() - new Date(observedAt).getTime()) / 60000))
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes} min ago`
  return `${Math.floor(minutes / 60)} hr ago`
}

export default function App() {
  const [pickup, setPickup] = useState('KL Sentral')
  const [dropoff, setDropoff] = useState('Pavilion Kuala Lumpur')
  const [route, setRoute] = useState({ pickup: 'KL Sentral', dropoff: 'Pavilion Kuala Lumpur' })
  const [quotes, setQuotes] = useState<Quote[]>([])
  const [provider, setProvider] = useState('Grab')
  const [service, setService] = useState('JustGrab')
  const [price, setPrice] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const loadQuotes = useCallback(async (nextRoute: { pickup: string; dropoff: string }) => {
    const query = new URLSearchParams(nextRoute)
    const response = await fetch(`/api/quotes?${query}`)
    if (!response.ok) throw new Error('Could not load fare observations.')
    const data: QuoteResponse = await response.json()
    setQuotes(data.quotes)
  }, [])

  useEffect(() => {
    loadQuotes(route).catch((cause) => setError(cause instanceof Error ? cause.message : 'Could not load quotes.'))
  }, [loadQuotes, route])

  async function compare(event: FormEvent) {
    event.preventDefault()
    setLoading(true)
    setError('')
    setMessage('')
    const nextRoute = { pickup: pickup.trim(), dropoff: dropoff.trim() }
    try {
      await loadQuotes(nextRoute)
      setRoute(nextRoute)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Could not load quotes.')
    } finally {
      setLoading(false)
    }
  }

  async function submitQuote(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    setError('')
    setMessage('')
    try {
      const response = await fetch('/api/quotes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pickup: route.pickup, dropoff: route.dropoff, provider, service, price: Number(price), observedAt: new Date().toISOString() }),
      })
      if (!response.ok) {
        const body = await response.json()
        throw new Error(body.error || 'Could not save the fare.')
      }
      await loadQuotes(route)
      setPrice('')
      setMessage('Fare saved. It will remain available after a refresh.')
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Could not save the fare.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <main>
      <nav><a className="brand" href="#">farelah<span>.</span></a><span className="badge">Klang Valley beta</span></nav>

      <section className="hero">
        <p className="eyebrow">One journey. Every option.</p>
        <h1>Find a better ride,<br /><em>before you book.</em></h1>
        <p className="intro">Compare fares that people have observed in their e-hailing apps.</p>

        <form className="route-form" onSubmit={compare}>
          <label><span>Pickup</span><input value={pickup} onChange={(e) => setPickup(e.target.value)} required maxLength={120} /></label>
          <div className="route-line" aria-hidden="true"><i /><i /></div>
          <label><span>Destination</span><input value={dropoff} onChange={(e) => setDropoff(e.target.value)} required maxLength={120} /></label>
          <button disabled={loading}>{loading ? 'Checking rides…' : 'Compare rides'} <b>→</b></button>
        </form>
        {error && <p className="error" role="alert">{error}</p>}
        {message && <p className="success" role="status">{message}</p>}
      </section>

      <section className="results">
        <header><div><p className="eyebrow">Community observations</p><h2>Ride options</h2></div><span>{route.pickup} → {route.dropoff}</span></header>
        {quotes.length ? <div className="cards">
          {quotes.map((quote, index) => {
            const stale = Date.now() - new Date(quote.observedAt).getTime() > 30 * 60 * 1000
            return <article key={quote.id} className={index === 0 ? 'best' : ''}>
              <div className="rank">{index === 0 ? 'Lowest observed' : `${index + 1}`}</div>
              <div><h3>{quote.provider}</h3><p>{quote.service}</p></div>
              <div className="time"><strong>{ageLabel(quote.observedAt)}</strong><span>{stale ? 'Older quote' : 'Recent quote'} · User submitted</span></div>
              <div className="price"><small>{quote.currency}</small>{quote.price.toFixed(2)}</div>
            </article>
          })}
        </div> : <p className="empty">No fares have been submitted for this route yet. Add one you see in a provider’s app below.</p>}
        <p className="notice">These are user submitted observations from the last 24 hours, not live provider quotes. Final fares and availability may differ.</p>

        <div className="contribute">
          <div><p className="eyebrow">Help improve the comparison</p><h2>Add a fare</h2><p>Enter a price you just saw for {route.pickup} → {route.dropoff}. Each provider’s latest observation appears above.</p></div>
          <form className="quote-form" onSubmit={submitQuote}>
            <label><span>Provider</span><select value={provider} onChange={(e) => setProvider(e.target.value)}><option>Grab</option><option>AirAsia Ride</option><option>Maxim</option><option>inDrive</option><option>Other</option></select></label>
            <label><span>Service</span><input value={service} onChange={(e) => setService(e.target.value)} required maxLength={120} placeholder="e.g. JustGrab" /></label>
            <label><span>Fare (MYR)</span><input type="number" min="0.01" max="10000" step="0.01" value={price} onChange={(e) => setPrice(e.target.value)} required placeholder="18.50" /></label>
            <button disabled={saving}>{saving ? 'Saving…' : 'Save fare'}</button>
          </form>
        </div>
      </section>
    </main>
  )
}
