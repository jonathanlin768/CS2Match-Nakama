import { useState } from "react"

const TOTAL_DOTS = 5

export default function CenterStage() {
  const [active, setActive] = useState(1)

  return (
    <div className="court-bg relative h-full flex-1 overflow-hidden rounded-md ring-1 ring-white/5">
      {/* faint arena arc */}
      <div className="pointer-events-none absolute right-6 top-1/2 h-72 w-72 -translate-y-1/2 rounded-full border-2 border-white/5" />

      {/* Title */}
      <div className="absolute left-6 top-5 z-20">
        <h2 className="text-4xl font-black tracking-wide text-foreground drop-shadow-lg">阵容</h2>
        <p className="mt-1 font-display text-3xl font-bold leading-none tracking-wider text-white/10">LINEUP</p>
        <p className="font-display text-sm tracking-[0.3em] text-white/10">ELITE</p>
      </div>

      {/* Player image */}
      <img
        src="/images/star-player.png"
        alt="当前出战选手"
        className="absolute bottom-0 left-1/2 z-10 h-[92%] -translate-x-1/2 object-contain drop-shadow-2xl"
      />

      {/* Empty roster slots on the right side */}
      <div className="absolute right-8 top-1/2 z-20 flex -translate-y-1/2 flex-col gap-6">
        {[0, 1, 2].map((i) => (
          <button
            key={i}
            type="button"
            aria-label={`空位 ${i + 1}`}
            className="grid h-12 w-12 place-items-center rounded-full bg-white/5 ring-1 ring-white/10 transition hover:bg-white/10"
          >
            <span className="text-2xl text-white/20">+</span>
          </button>
        ))}
      </div>

      {/* Pagination dots */}
      <div className="absolute bottom-4 left-1/2 z-20 flex -translate-x-1/2 gap-2">
        {Array.from({ length: TOTAL_DOTS }).map((_, i) => (
          <button
            key={i}
            type="button"
            aria-label={`切换到第 ${i + 1} 套阵容`}
            onClick={() => setActive(i)}
            className={`h-2 rounded-full transition-all ${
              i === active ? "w-6 bg-gold" : "w-2 bg-white/25"
            }`}
          />
        ))}
      </div>

      {/* Bottom watermark */}
      <span className="pointer-events-none absolute bottom-1 left-1/2 -translate-x-1/2 font-display text-xs tracking-[0.6em] text-white/5">
        CS2 SIMU
      </span>
    </div>
  )
}
