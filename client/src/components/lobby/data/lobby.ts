import {
  Wallet,
  Ticket,
  TicketPercent,
  PackageOpen,
  Gift,
  Target,
  Store,
  ShoppingCart,
  FileText,
  ArrowLeftRight,
  Shield,
  Trophy,
  Building2,
  Users,
  Boxes,
  type LucideIcon,
} from "lucide-react"

export interface PromoItem {
  label: string
  icon: LucideIcon
  badge?: string
  hasGift?: boolean
}

export const promoItems: PromoItem[] = [
  { label: "新手启航", icon: Wallet },
  { label: "代金券", icon: TicketPercent, badge: "0.1折" },
  { label: "膨胀返券", icon: Ticket },
  { label: "十万红包", icon: PackageOpen },
  { label: "超值好礼", icon: Gift, hasGift: true },
  { label: "七日目标", icon: Target, hasGift: true },
  { label: "特惠商城", icon: Store, hasGift: true },
]

export interface NavItem {
  label: string
  icon: LucideIcon
  locked?: boolean
  dot?: boolean
}

export const bottomNav: NavItem[] = [
  { label: "商店", icon: ShoppingCart, locked: true },
  { label: "合同", icon: FileText, locked: true },
  { label: "交易", icon: ArrowLeftRight, locked: true },
  { label: "联盟", icon: Shield, locked: true },
  { label: "荣誉室", icon: Trophy },
  { label: "办公室", icon: Building2 },
  { label: "选手", icon: Users, dot: true },
  { label: "仓库", icon: Boxes },
]

export interface ModeCard {
  title: string
  subtitle: string
  hint?: string
  locked?: boolean
}

export const rightModes: ModeCard[] = [
  { title: "竞技赛", subtitle: "COMPETITIVE MATCH", hint: "通关S1-16开启", locked: true },
  { title: "商业赛", subtitle: "COMMERCIAL MATCH", hint: "通关S1-28开启", locked: true },
  { title: "赛季联赛", subtitle: "CHALLENGE MATCH", hint: "开服第12天开启", locked: true },
]
