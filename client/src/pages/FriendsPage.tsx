import type { Friend } from "@heroiclabs/nakama-js"
import { Check, ChevronLeft, ContactRound, Plus, RefreshCw, UserPlus, X } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import { addFriendById, addFriendsByUsername, deleteFriend, listFriends } from "../api/friends"
import { getContactExchange, requestContactExchange, respondContactExchange, setContactProfile, type ContactExchange, type ContactProfileSummary } from "../api/social"
import { useAuth } from "../context/AuthContext"
import { useContactInbox, type ContactInboxSummary } from "../hooks/useContactInbox"
import { subscribeNotifications, useSocket } from "../hooks/useSocket"
import { socialContactExchangeEnabled } from "../features"
import { groupFriends } from "../social/friend-groups"

function code(username?: string) { return `玩家#${username ?? "--------"}` }

export default function FriendsPage() {
  const navigate = useNavigate()
  const { session, isGuest } = useAuth()
  const [friends, setFriends] = useState<Friend[]>([])
  const [loading, setLoading] = useState(true)
  const [addCode, setAddCode] = useState("")
  const [selected, setSelected] = useState<Friend | null>(null)
  const [exchange, setExchange] = useState<ContactExchange | null>(null)
  const [qq, setQQ] = useState("")
  const [wechat, setWeChat] = useState("")
  const [profileSummary, setProfileSummary] = useState<ContactProfileSummary | null>(null)
  const friendsSocket = useSocket(session)
  const { accepted, incoming, outgoing } = useMemo(() => groupFriends(friends), [friends])
  const friendIds = useMemo(() => accepted.flatMap((friend) => friend.user?.id ? [friend.user.id] : []), [accepted])
  const inbox = useContactInbox(session, friendIds, selected?.user?.id)
  const cards = selected?.user?.id ? inbox.cardsByFriend[selected.user.id] ?? [] : []

  const load = useCallback(async () => {
    if (!session) return
    setLoading(true)
    try { setFriends((await listFriends(session)).friends ?? []) }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
    finally { setLoading(false) }
  }, [session])
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0)
    return () => window.clearTimeout(timer)
  }, [load])
  useEffect(() => {
    if (!friendsSocket) return
    return subscribeNotifications(() => { void load() })
  }, [friendsSocket, load])
  useEffect(() => {
    const friendId = selected?.user?.id
    if (!session || !friendId) return
    let active = true
    void getContactExchange(session, friendId)
      .then((value) => { if (active) setExchange(value) })
      .catch((error) => { if (active) toast.error(error instanceof Error ? error.message : String(error)) })
    return () => { active = false }
  }, [session, selected?.user?.id, inbox.revision])

  async function add() {
    if (!session) return
    const username = addCode.trim().replace(/^玩家#/i, "").toUpperCase()
    if (!/^[A-Z0-9]{8}$/.test(username)) { toast.info("请输入 8 位玩家代码"); return }
    try { await addFriendsByUsername(session, username); setAddCode(""); await load(); toast.success("好友请求已发送") }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
  }

  async function openExchange(friend: Friend) {
    if (!friend.user?.id) return
    setSelected(friend); setExchange(null)
  }

  async function saveProfile() {
    if (!session) return
    try { setProfileSummary(await setContactProfile(session, { qq, wechat })); toast.success("联系方式已私密保存") }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
  }

  async function requestExchange() {
    if (!session || !selected?.user?.id) return
    try { setExchange(await requestContactExchange(session, selected.user.id, ["qq", "wechat"])); toast.success("交换请求已发送") }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
  }

  async function respond(accept: boolean) {
    if (!session || !selected?.user?.id || !exchange?.request_id) return
    try { setExchange(await respondContactExchange(session, selected.user.id, exchange.request_id, accept)); toast.success(accept ? "已互相授权联系方式" : "已拒绝请求") }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
  }

  if (isGuest) return <div className="page-center"><ContactRound size={38} /><h1>登录后使用好友功能</h1><p>游客可以体验教学战和电脑对战。请使用右上角“登录”绑定邮箱或切换已有账号，再添加好友和交换联系方式。</p><button className="secondary-button" onClick={() => navigate("/")}>返回主页</button></div>

  return <div className="sub-page friends-page">
    <button className="back-button" onClick={() => navigate("/")}><ChevronLeft size={18} />返回主页</button>
    <header className="page-heading friends-heading"><div><p className="eyebrow">FRIENDS</p><h1>好友与联系方式</h1><p>只能通过 8 位玩家代码添加好友；没有自由聊天，只能发送交换联系方式卡片。</p></div><button className="icon-button" onClick={() => void load()}><RefreshCw size={18} />刷新</button></header>
    <section className="friend-add"><input value={addCode} maxLength={11} onChange={(event) => setAddCode(event.target.value)} placeholder="玩家#1234ABCD" /><button className="primary-button" onClick={() => void add()}><UserPlus size={18} />添加好友</button></section>
    <div className="friends-layout"><section className="friends-list">
      {incoming.length > 0 && <FriendGroup title="收到的请求" friends={incoming} action={(friend) => <div className="friend-mini-actions"><button onClick={() => { if (session && friend.user?.id) void addFriendById(session, friend.user.id).then(load) }}><Check size={16} /></button><button onClick={() => { if (session && friend.user?.id) void deleteFriend(session, friend.user.id).then(load) }}><X size={16} /></button></div>} />}
      <FriendGroup title={`我的好友 · ${accepted.length}`} friends={accepted} summaries={inbox.summaries} action={socialContactExchangeEnabled ? (friend) => <button className="exchange-button" onClick={() => void openExchange(friend)}><ContactRound size={17} />联系方式</button> : undefined} empty={loading ? "正在加载…" : "还没有好友，使用上方玩家代码添加"} />
      {outgoing.length > 0 && <FriendGroup title="已发出的请求" friends={outgoing} />}
    </section>{socialContactExchangeEnabled && <aside className="contact-panel">
      <h2>我的联系方式</h2><p>这些内容不会公开，也不会写进聊天卡片。只有双方接受交换后才能互相读取。</p>{profileSummary && <small className="profile-summary">QQ {profileSummary.qq_masked || "未设置"} · 微信 {profileSummary.wechat_masked || "未设置"}</small>}<label>QQ<input value={qq} onChange={(event) => setQQ(event.target.value)} placeholder="5–12 位数字" /></label><label>微信号<input value={wechat} onChange={(event) => setWeChat(event.target.value)} placeholder="6–20 位字母、数字或 _ -" /></label><button className="secondary-button" onClick={() => void saveProfile()}>私密保存</button>
    </aside>}</div>
    {selected && <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelected(null) }}><section className="modal-card exchange-modal"><button className="modal-close" onClick={() => setSelected(null)}><X size={20} /></button><p className="eyebrow">CONTACT EXCHANGE</p><h2>与 {code(selected.user?.username)} 交换联系方式</h2>{cards.length > 0 && <div className="card-history">{cards.slice(-3).map((card) => <small key={card.messageId}>卡片 · {card.action === "requested" ? "已发起" : card.action === "accepted" ? "已接受" : "已拒绝"} · v{card.version}</small>)}</div>}{!exchange ? <p>正在读取权威状态…</p> : <ExchangeCard exchange={exchange} me={session?.user_id ?? ""} onRequest={requestExchange} onRespond={respond} />}</section></div>}
  </div>
}

function FriendGroup({ title, friends, action, empty, summaries }: { title: string; friends: Friend[]; action?: (friend: Friend) => React.ReactNode; empty?: string; summaries?: Record<string, ContactInboxSummary> }) {
  return <div className="friend-group"><h2>{title}</h2>{friends.length === 0 ? <p className="empty-copy">{empty}</p> : friends.map((friend) => { const summary = friend.user?.id ? summaries?.[friend.user.id] : undefined; return <article className="friend-row" key={friend.user?.id}><span className="friend-avatar">{friend.user?.username?.slice(0, 2) ?? "??"}</span><div><b>{code(friend.user?.username)}</b><small>{friend.user?.online ? "在线" : "离线"}{summary?.latest ? ` · ${summary.latest.action === "requested" ? "请求交换联系方式" : summary.latest.action === "accepted" ? "已接受交换" : "已拒绝交换"}` : ""}</small></div>{summary && summary.unread > 0 && <span className="unread-badge" aria-label={`${summary.unread} 条未读交换卡片`}>{summary.unread}</span>}{action?.(friend)}</article> })}</div>
}

function ExchangeCard({ exchange, me, onRequest, onRespond }: { exchange: ContactExchange; me: string; onRequest: () => Promise<void>; onRespond: (accept: boolean) => Promise<void> }) {
  if (exchange.status === "none" || exchange.status === "declined" || exchange.status === "expired") return <div className="exchange-state"><ContactRound size={28} /><p>向好友发送一张结构化请求卡片。请求中不会包含 QQ 或微信正文。</p><button className="primary-button" onClick={() => void onRequest()}><Plus size={17} />请求交换 QQ 和微信</button></div>
  if (exchange.status === "pending") {
    const canRespond = exchange.recipient_id === me
    return <div className="exchange-state pending"><b>交换请求等待处理</b><p>请求渠道：{exchange.channels?.join("、")}</p>{canRespond ? <div className="modal-actions"><button className="secondary-button" onClick={() => void onRespond(false)}>拒绝</button><button className="primary-button" onClick={() => void onRespond(true)}>接受并互相授权</button></div> : <small>好友接受后，双方才能看到联系方式。</small>}</div>
  }
  return <div className="exchange-state accepted"><b><Check size={18} />已互相授权</b><p>好友 QQ：<strong>{exchange.friend_contact?.qq || "未提供"}</strong></p><p>好友微信：<strong>{exchange.friend_contact?.wechat || "未提供"}</strong></p><small>如果你们不再是好友，服务器会立即停止返回这些内容。</small></div>
}
