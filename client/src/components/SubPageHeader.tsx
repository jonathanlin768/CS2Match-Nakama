import { ArrowLeft } from "lucide-react"
import { useNavigate } from "react-router-dom"

type SubPageHeaderProps = {
  /** 标题文本，如 "匹配比赛" */
  title: string
  /** 自定义返回行为；默认返回浏览器上一页 */
  onBack?: () => void
}

/**
 * 二级界面通用头部：左上角「← 标题」，点击返回。
 * 各二级页面在固定 1920x900 画框顶部放置 <SubPageHeader title="..." /> 即可复用。
 */
export default function SubPageHeader({ title, onBack }: SubPageHeaderProps) {
  const navigate = useNavigate()
  const handleBack = onBack ?? (() => navigate(-1))

  return (
    <header className="flex h-[88px] shrink-0 items-center px-[40px]">
      <button
        type="button"
        onClick={handleBack}
        aria-label={`返回（${title}）`}
        className="group flex items-center gap-3 rounded-md py-2 pr-4 text-foreground transition hover:text-gold active:scale-[0.98]"
      >
        <ArrowLeft size={32} strokeWidth={2.5} className="transition-transform group-hover:-translate-x-1" />
        <span className="text-3xl font-black tracking-wide">{title}</span>
      </button>
    </header>
  )
}
