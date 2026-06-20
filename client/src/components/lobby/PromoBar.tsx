import { Gift } from "lucide-react"
import { promoItems } from "./data/lobby"

export default function PromoBar() {
  return (
    <div className="flex h-[100px] items-end justify-end gap-6 px-4">
      {promoItems.map((item) => {
        const Icon = item.icon
        return (
          <button
            key={item.label}
            type="button"
            className="group relative flex w-16 flex-col items-center gap-1 transition active:scale-95"
          >
            <div className="relative grid h-12 w-12 place-items-center rounded-lg bg-gradient-to-b from-panel-light to-panel ring-1 ring-white/10 transition group-hover:ring-gold/40">
              <Icon size={26} className="text-gold" />
              {item.badge && (
                <span className="absolute -left-1 -top-2 rounded-sm bg-accent px-1 py-0.5 text-[10px] font-bold leading-none text-accent-foreground">
                  {item.badge}
                </span>
              )}
              {item.hasGift && (
                <span className="absolute -right-1.5 -top-1.5 grid h-4 w-4 place-items-center rounded-full bg-accent">
                  <Gift size={10} className="text-accent-foreground" />
                </span>
              )}
            </div>
            <span className="text-xs font-medium text-foreground/90">{item.label}</span>
          </button>
        )
      })}
    </div>
  )
}
