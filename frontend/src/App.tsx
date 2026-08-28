import { FormEvent, useState } from 'react'

type Quote = {
  provider: string
  service: string
  price: number
  currency: string
  pickupMinutes: number
  journeyMinutes: number
  isSample: boolean
}

type QuoteResponse = {
  pickup: string
  dropoff: string
  observedAt: string
  quotes: Quote[]
}

const previewQuotes: Quote[] = [
  { provider: 'Maxim', service: 'Economy', price: 16.53, currency: 'MYR', pickupMinutes: 9, journeyMinutes: 20, isSample: true },
  { provider: 'AirAsia Ride', service: 'Compact', price: 17.83, currency: 'MYR', pickupMinutes: 7, journeyMinutes: 19, isSample: true },
  { provider: 'Grab', service: 'JustGrab', price: 20.23, currency: 'MYR', pickupMinutes: 4, journeyMinutes: 18, isSample: true },
]

export default function App() {
  const [pickup, setPickup] = useState('KL Sentral')
  const [dropoff, setDropoff] = useState('Pavilion Kuala Lumpur')
  const [quotes, setQuotes] = useState(previewQuotes)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [compared, setCompared] = useState(false)

  async function compare(event: FormEvent) {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      const query = new URLSearchParams({ pickup, dropoff })
      const response = await fetch(`/api/quotes?${query}`)
      if (!response.ok) throw new Error('The comparison service is not available yet.')
      const data: QuoteResponse = await response.json()
      setQuotes([...data.quotes].sort((a, b) => a.price - b.price))
      setCompared(true)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to compare fares.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main>
      <nav><a className="brand" href="#">farelah<span>.</span></a><span className="badge">Klang Valley beta</span></nav>

      <section className="hero">
        <p className="eyebrow">One journey. Every option.</p>
        <h1>Find a better ride,<br /><em>before you book.</em></h1>
        <p className="intro">Compare indicative e-hailing fares in one calm, simple view.</p>

        <form onSubmit={compare}>
          <label><span>Pickup</span><input value={pickup} onChange={(e) => setPickup(e.target.value)} required /></label>
          <div className="route-line" aria-hidden="true"><i /><i /></div>
          <label><span>Destination</span><input value={dropoff} onChange={(e) => setDropoff(e.target.value)} required /></label>
          <button disabled={loading}>{loading ? 'Checking rides…' : 'Compare rides'} <b>→</b></button>
        </form>
        {error && <p className="error">{error} Start the Go API to enable live requests.</p>}
      </section>

      <section className="results">
        <header><div><p className="eyebrow">{compared ? 'Comparison ready' : 'Example comparison'}</p><h2>Ride options</h2></div><span>Indicative sample fares</span></header>
        <div className="cards">
          {quotes.map((quote, index) => (
            <article key={quote.provider} className={index === 0 ? 'best' : ''}>
              <div className="rank">{index === 0 ? 'Best value' : `${index + 1}`}</div>
              <div><h3>{quote.provider}</h3><p>{quote.service}</p></div>
              <div className="time"><strong>{quote.pickupMinutes} min</strong><span>pickup · {quote.journeyMinutes} min ride</span></div>
              <div className="price"><small>{quote.currency}</small>{quote.price.toFixed(2)}</div>
            </article>
          ))}
        </div>
        <p className="notice">Sample data only—not live provider prices. Final fares, tolls and availability may differ in each provider's official app.</p>
      </section>
    </main>
  )
}
