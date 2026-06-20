import { Plus, Users, Mail, Gem, Lock, BarChart3, BookOpen, Hand } from "lucide-react"

function CurrencyPill({ icon, value }: { icon: React.ReactNode; value: string }) {
  return (
    <div className="flex items-center gap-2 rounded-full bg-black/50 pl-1 pr-1 py-1 ring-1 ring-white/10">
      {icon}
      <span className="font-display text-lg font-semibold tabular-nums text-foreground">{value}</span>
      <button
        type="button"
        className="grid h-6 w-6 place-items-center rounded-full bg-accent text-accent-foreground shadow-md transition active:scale-90"
        aria-label="增加"
      >
        <Plus size={16} strokeWidth={3} />
      </button>
    </div>
  )
}

function IconButton({ label, children, locked }: { label: string; children: React.ReactNode; locked?: boolean }) {
  return (
    <button
      type="button"
      aria-label={label}
      className="relative grid h-10 w-10 place-items-center rounded-md text-foreground/80 transition hover:bg-white/5 active:scale-95"
    >
      {children}
      {locked && (
        <span className="absolute -right-0.5 -top-0.5 grid h-4 w-4 place-items-center rounded-full bg-black/70">
          <Lock size={10} className="text-gold" />
        </span>
      )}
    </button>
  )
}

export default function TopBar() {
  return (
    <header className="flex h-[80px] items-center justify-between gap-4 px-4">
      {/* Player profile */}
      <div className="flex items-center gap-3">
        <div className="relative h-14 w-14 shrink-0 overflow-hidden rounded-lg ring-2 ring-gold/70">
          <img
            src="/images/star-player.png"
            alt="玩家头像"
            className="h-full w-full scale-150 object-cover object-top"
          />
        </div>
        <div className="min-w-[180px]">
          <div className="flex items-center gap-2">
            <span className="font-display text-xl font-bold text-gold">LV.5</span>
            <span className="text-white/30">|</span>
            <span className="text-base font-bold tracking-wide">指挥官</span>
          </div>
          <div className="mt-1 flex items-center gap-3 text-sm text-muted">
            <span className="flex items-center gap-1">
              <span className="font-display font-bold text-accent">战</span>
              <span className="font-display font-semibold text-foreground">4881</span>
            </span>
            <Hand size={16} className="text-white/40" />
            <BookOpen size={16} className="text-white/40" />
            <span className="flex items-center gap-1">
              <BookOpen size={16} className="text-white/40" />
              <span className="text-xs text-gold">64%</span>
            </span>
          </div>
        </div>
      </div>

      {/* System icons */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-3">
        <CurrencyPill icon={<Gem size={20} className="text-cyan-300" />} value="1100" />
        <CurrencyPill icon={<div className="h-5 w-7 rounded-sm bg-gradient-to-b from-pink-400 to-rose-600" />} value="0" />
      </div>
        <IconButton label="好友">
          <Users size={22} />
        </IconButton>
        <IconButton label="邮件">
          <Mail size={22} />
        </IconButton>
        <IconButton label="个人信息" locked>
          <Users size={22} />
        </IconButton>
        <IconButton label="排行榜" locked>
          <BarChart3 size={22} />
        </IconButton>
      </div>
    </header>
  )
}
