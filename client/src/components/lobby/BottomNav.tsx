import { Lock } from "lucide-react"
import { bottomNav } from "./data/lobby"

export default function BottomNav() {
  return (
    <nav className="flex h-[130px] items-center justify-center gap-2 bg-gradient-to-t from-black/80 to-transparent px-4">
      {bottomNav.map((item, idx) => {
        const Icon = item.icon
        return (
          <div key={item.label} className="flex items-center">
            <button
              type="button"
              disabled={item.locked}
              className="group relative flex items-center gap-3 rounded-md px-5 py-2 transition hover:bg-white/5 active:scale-95 disabled:opacity-60"
            >
              <div className="relative">
                <Icon size={36} className={item.locked ? "text-muted" : "text-foreground"} />
                {item.dot && (
                  <span className="absolute -right-1 -top-1 h-3 w-3 rounded-full bg-gold shadow-[0_0_6px] shadow-gold" />
                )}
              </div>
              <span className={`text-xl font-bold ${item.locked ? "text-muted" : "text-foreground"}`}>
                {item.label}
              </span>
              {item.locked && <Lock size={18} className="text-gold" />}
            </button>
            {idx < bottomNav.length - 1 && <span className="h-8 w-px bg-white/10" />}
          </div>
        )
      })}
    </nav>
  )
}
