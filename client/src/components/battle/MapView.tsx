import { Map } from "lucide-react"

/**
 * 地图区域占位。后续接入实时对局数据后，这里渲染类似 HLTV 的雷达地图
 * （选手位置、炸弹点、存活状态等）。当前留出区域并标注。
 */
export default function MapView() {
  return (
    <div className="court-bg relative flex min-h-0 flex-1 flex-col items-center justify-center overflow-hidden rounded-md bg-panel/50 ring-1 ring-white/10">
      <Map size={56} className="text-white/15" strokeWidth={1.5} />
      <span className="mt-3 font-display text-sm tracking-[0.35em] text-white/20">MAP · 地图区域</span>
    </div>
  )
}
