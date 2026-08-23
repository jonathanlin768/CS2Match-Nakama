import type { Friend } from "@heroiclabs/nakama-js"
import { AlertTriangle, Check, ChevronLeft, ContactRound, Plus, RefreshCw, RotateCcw, UserPlus, X } from "lucide-react"
import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import { addFriendById, addFriendsByUsername, deleteFriend, listFriends } from "../api/friends"
import {
  getContactExchange,
  getContactProfile,
  requestContactExchange,
  respondContactExchange,
  setContactProfile,
  socialErrorMessage,
  type ContactExchange,
  type ContactExchangeInbox,
  type ContactInboxItem,
  type ContactProfileView,
} from "../api/social"
import { useAuth } from "../context/AuthContext"
import { useContactExchangeInbox } from "../context/ContactExchangeContext"
import { useContactInbox, type ContactInboxSummary } from "../hooks/useContactInbox"
import { subscribeNotifications, useSocket } from "../hooks/useSocket"
import { socialContactExchangeEnabled } from "../features"
import { contactChannels, contactItemExpired, contactProfileChanged, disclosedFriendContact, formatContactTime } from "../social/contact-inbox"
import { groupFriends } from "../social/friend-groups"

interface SelectedContact { id: string; username?: string }

function code(username?: string) { return `玩家#${username ?? "--------"}` }
function channelLabel(channels?: string[]) { return channels?.map((channel) => channel === "qq" ? "QQ" : channel === "wechat" ? "微信" : channel).join("、") || "未指定" }
function eventLabel(action: string) {
  return ({ requested: "已发起", accepted: "已接受", declined: "已拒绝", stale: "已失效", revoked: "已撤销" } as Record<string, string>)[action] ?? "状态更新"
}

export default function FriendsPage() {
  const navigate = useNavigate()
  const { session, isGuest } = useAuth()
  const authority = useContactExchangeInbox()
  const [friends, setFriends] = useState<Friend[]>([])
  const [loading, setLoading] = useState(true)
  const [addCode, setAddCode] = useState("")
  const [selected, setSelected] = useState<SelectedContact | null>(null)
  const [exchange, setExchange] = useState<ContactExchange | null>(null)
  const [exchangeError, setExchangeError] = useState<string | null>(null)
  const [actionBusy, setActionBusy] = useState<string | null>(null)
  const [qq, setQQ] = useState("")
  const [wechat, setWeChat] = useState("")
  const [profile, setProfile] = useState<ContactProfileView | null>(null)
  const [profileStatus, setProfileStatus] = useState<"idle" | "loading" | "success" | "error">("idle")
  const [profileError, setProfileError] = useState<string | null>(null)
  const profileTouched = useRef(false)
  const friendsSocket = useSocket(session)
  const { accepted, incoming, outgoing } = useMemo(() => groupFriends(friends), [friends])
  const friendIds = useMemo(() => accepted.flatMap((friend) => friend.user?.id ? [friend.user.id] : []), [accepted])
  const cardInbox = useContactInbox(session, friendIds, selected?.id)
  const cards = selected ? cardInbox.cardsByFriend[selected.id] ?? [] : []

  const load = useCallback(async () => {
    if (!session) return
    setLoading(true)
    try { setFriends((await listFriends(session)).friends ?? []) }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
    finally { setLoading(false) }
  }, [session])

  const loadProfile = useCallback(async () => {
    if (!session || !socialContactExchangeEnabled) return
    setProfileStatus("loading")
    try {
      const value = await getContactProfile(session)
      setProfile(value)
      if (!profileTouched.current) {
        setQQ(value.qq ?? "")
        setWeChat(value.wechat ?? "")
      }
      setProfileError(null)
      setProfileStatus("success")
    } catch (error) {
      setProfileError(socialErrorMessage(error))
      setProfileStatus("error")
    }
  }, [session])

  useEffect(() => {
    const timer = window.setTimeout(() => { void load(); void loadProfile() }, 0)
    return () => window.clearTimeout(timer)
  }, [load, loadProfile])
  useEffect(() => {
    if (!friendsSocket) return
    return subscribeNotifications(() => { void load() })
  }, [friendsSocket, load])
  useEffect(() => {
    const friendId = selected?.id
    if (!session || !friendId) return
    let active = true
    void getContactExchange(session, friendId)
      .then((value) => { if (active) { setExchange(value); setExchangeError(null) } })
      .catch((error) => { if (active) { setExchangeError(socialErrorMessage(error)); setExchange(null) } })
    return () => { active = false }
  }, [session, selected?.id, cardInbox.revision, authority.revision])

  async function refreshAll() {
    await Promise.allSettled([load(), loadProfile(), authority.refresh()])
  }

  async function add() {
    if (!session) return
    const username = addCode.trim().replace(/^玩家#/i, "").toUpperCase()
    if (!/^[A-Z0-9]{8}$/.test(username)) { toast.info("请输入 8 位玩家代码"); return }
    try { await addFriendsByUsername(session, username); setAddCode(""); await load(); toast.success("好友请求已发送") }
    catch (error) { toast.error(error instanceof Error ? error.message : String(error)) }
  }

  function openExchange(friend: Friend | ContactInboxItem) {
    const id = "friend_id" in friend ? friend.friend_id : friend.user?.id
    const username = "friend_id" in friend ? friend.username : friend.user?.username
    if (!id) return
    setSelected({ id, username }); setExchange(null); setExchangeError(null)
  }

  async function saveProfile() {
    if (!session) return
    const changed = contactProfileChanged(profile, qq, wechat)
    if (changed && (profile?.revision ?? 0) > 0 && !window.confirm("修改或清空任一联系方式会让所有既有交换授权立即失效，双方需要重新申请并接受。仍要保存吗？")) return
    try {
      const value = await setContactProfile(session, { qq, wechat })
      setProfile(value)
      setQQ(value.qq ?? "")
      setWeChat(value.wechat ?? "")
      profileTouched.current = false
      setProfileError(null)
      setProfileStatus("success")
      await authority.refresh()
      toast.success(changed ? "联系方式已保存；既有授权状态已刷新" : "联系方式没有变化")
    } catch (error) { toast.error(socialErrorMessage(error)) }
  }

  async function requestExchange(target = selected) {
    if (!session || !target) return
    const channels = contactChannels(profile ?? {})
    if (channels.length === 0) { toast.info("请先保存至少一种联系方式"); return }
    setActionBusy(`request:${target.id}`)
    try {
      const value = await requestContactExchange(session, target.id, channels)
      if (selected?.id === target.id) setExchange(value)
      await authority.refresh()
      toast.success("交换请求已发送；即使消息卡片未送达，好友也能在申请列表中看到")
    } catch (error) {
      toast.error(socialErrorMessage(error))
      await authority.refresh()
    } finally { setActionBusy(null) }
  }

  async function respondItem(item: ContactInboxItem, accept: boolean) {
    if (!session || !item.request_id) return
    setActionBusy(`${accept ? "accept" : "decline"}:${item.request_id}`)
    try {
      const value = await respondContactExchange(session, item.friend_id, item.request_id, accept)
      if (selected?.id === item.friend_id) setExchange(value)
      await authority.refresh()
      toast.success(accept ? "已互相授权联系方式" : "已拒绝请求")
    } catch (error) {
      toast.error(socialErrorMessage(error))
      await authority.refresh()
      if (selected?.id === item.friend_id) setExchange(null)
    } finally { setActionBusy(null) }
  }

  async function respondSelected(accept: boolean) {
    if (!selected || !exchange?.request_id) return
    await respondItem({
      friend_id: selected.id,
      username: selected.username ?? "",
      request_id: exchange.request_id,
      requester_id: exchange.requester_id,
      recipient_id: exchange.recipient_id,
      channels: exchange.channels,
      status: exchange.status,
    }, accept)
  }

  if (isGuest) return <div className="page-center"><ContactRound size={38} /><h1>登录后使用好友功能</h1><p>游客可以体验教学战和电脑对战。请使用右上角“登录”绑定邮箱或切换已有账号，再添加好友和交换联系方式。</p><button className="secondary-button" onClick={() => navigate("/")}>返回主页</button></div>

  return <div className="sub-page friends-page">
    <button className="back-button" onClick={() => navigate("/")}><ChevronLeft size={18} />返回主页</button>
    <header className="page-heading friends-heading"><div><p className="eyebrow">FRIENDS</p><h1>好友与联系方式</h1><p>只能通过 8 位玩家代码添加好友；没有自由聊天，只能发送交换联系方式卡片。</p></div><button className="icon-button" onClick={() => void refreshAll()}><RefreshCw size={18} />刷新</button></header>
    <section className="friend-add"><input value={addCode} maxLength={11} onChange={(event) => setAddCode(event.target.value)} placeholder="玩家#1234ABCD" /><button className="primary-button" onClick={() => void add()}><UserPlus size={18} />添加好友</button></section>
    <div className="friends-layout"><section className="friends-list">
      {socialContactExchangeEnabled && <ContactRequestsSection inbox={authority.inbox} loading={authority.status === "loading"} error={authority.error?.message} busy={actionBusy} onOpen={openExchange} onRespond={respondItem} onRequest={(item) => requestExchange({ id: item.friend_id, username: item.username })} onRetry={authority.refresh} />}
      {incoming.length > 0 && <FriendGroup title="收到的好友请求" friends={incoming} action={(friend) => <div className="friend-mini-actions"><button aria-label="接受好友请求" onClick={() => { if (session && friend.user?.id) void addFriendById(session, friend.user.id).then(load) }}><Check size={16} /></button><button aria-label="拒绝好友请求" onClick={() => { if (session && friend.user?.id) void deleteFriend(session, friend.user.id).then(load) }}><X size={16} /></button></div>} />}
      <FriendGroup title={`我的好友 · ${accepted.length}`} friends={accepted} summaries={cardInbox.summaries} action={socialContactExchangeEnabled ? (friend) => <button className="exchange-button" onClick={() => openExchange(friend)}><ContactRound size={17} />联系方式</button> : undefined} empty={loading ? "正在加载…" : "还没有好友，使用上方玩家代码添加"} />
      {outgoing.length > 0 && <FriendGroup title="已发出的好友请求" friends={outgoing} />}
    </section>{socialContactExchangeEnabled && <aside className="contact-panel">
      <h2>我的联系方式</h2><p>QQ 和微信均为选填，至少保存一种即可。内容只在双方授权且资料均未变化时展示；修改或清空任一项会让整次既有授权失效。</p>
      {profile && <small className="profile-summary">QQ {profile.summary.qq_masked || "未设置"} · 微信 {profile.summary.wechat_masked || "未设置"} · 版本 {profile.revision ?? 0}</small>}
      {profileStatus === "loading" && !profile && <p className="contact-inline-state">正在安全读取本人资料…</p>}
      {profileError && <div className="contact-inline-error" role="alert"><span>{profileError} 本地已编辑内容不会被清空。</span><button onClick={() => void loadProfile()}><RotateCcw size={14} />重试</button></div>}
      <label>QQ（选填）<input aria-label="QQ" value={qq} onChange={(event) => { profileTouched.current = true; setQQ(event.target.value) }} placeholder="留空可清除；5–12 位数字" disabled={profileStatus === "loading" && !profile} /></label>
      <label>微信号（选填）<input aria-label="微信号" value={wechat} onChange={(event) => { profileTouched.current = true; setWeChat(event.target.value) }} placeholder="留空可清除；6–20 位字母、数字或 _ -" disabled={profileStatus === "loading" && !profile} /></label>
      <button className="secondary-button" onClick={() => void saveProfile()} disabled={profileStatus === "loading" && !profile}>私密保存</button>
    </aside>}</div>
    {cardInbox.cardError && <div className="contact-history-error" role="status"><span>{cardInbox.cardError}</span><button onClick={cardInbox.retry}><RotateCcw size={14} />重试卡片历史</button></div>}
    {selected && <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelected(null) }}><section className="modal-card exchange-modal"><button className="modal-close" aria-label="关闭" onClick={() => setSelected(null)}><X size={20} /></button><p className="eyebrow">CONTACT EXCHANGE</p><h2>与 {code(selected.username)} 交换联系方式</h2>{cards.length > 0 && <div className="card-history" aria-label="交换事件历史">{cards.slice(-3).map((card) => <small key={card.messageId}>卡片 · {eventLabel(card.action)} · v{card.version}</small>)}</div>}{exchangeError ? <div className="exchange-state error" role="alert"><p>{exchangeError}</p><button className="secondary-button" onClick={() => { setExchangeError(null); setExchange(null); void authority.refresh() }}><RefreshCw size={15} />刷新状态</button></div> : !exchange ? <p>正在读取权威状态…</p> : <ExchangeCard exchange={exchange} me={session?.user_id ?? ""} busy={Boolean(actionBusy)} onRequest={() => requestExchange()} onRespond={respondSelected} />}</section></div>}
  </div>
}

function FriendGroup({ title, friends, action, empty, summaries }: { title: string; friends: Friend[]; action?: (friend: Friend) => ReactNode; empty?: string; summaries?: Record<string, ContactInboxSummary> }) {
  return <div className="friend-group"><h2>{title}</h2>{friends.length === 0 ? <p className="empty-copy">{empty}</p> : friends.map((friend) => { const summary = friend.user?.id ? summaries?.[friend.user.id] : undefined; return <article className="friend-row" key={friend.user?.id}><span className="friend-avatar">{friend.user?.username?.slice(0, 2) ?? "??"}</span><div><b>{code(friend.user?.username)}</b><small>{friend.user?.online ? "在线" : "离线"}{summary?.latest ? ` · ${eventLabel(summary.latest.action)}` : ""}</small></div>{summary && summary.unread > 0 && <span className="unread-badge" aria-label={`${summary.unread} 条未读交换卡片`}>{summary.unread}</span>}{action?.(friend)}</article> })}</div>
}

export function ContactRequestsSection({ inbox, loading, error, busy, onOpen, onRespond, onRequest, onRetry }: {
  inbox: ContactExchangeInbox
  loading?: boolean
  error?: string
  busy?: string | null
  onOpen: (item: ContactInboxItem) => void
  onRespond: (item: ContactInboxItem, accept: boolean) => Promise<void>
  onRequest: (item: ContactInboxItem) => Promise<void>
  onRetry: () => Promise<void>
}) {
  const total = inbox.received.length + inbox.sent.length + inbox.reapproval.length
  return <div className="friend-group contact-requests" id="contact-requests"><div className="contact-group-heading"><h2>联系方式申请{inbox.incoming_pending_count > 0 ? ` · ${inbox.incoming_pending_count} 待处理` : ""}</h2>{error && <button onClick={() => void onRetry()}><RotateCcw size={14} />重试</button>}</div>
    {error && <p className="contact-request-error" role="alert">{error}</p>}
    {!error && total === 0 && <p className="empty-copy">{loading ? "正在读取权威申请…" : "暂无联系方式申请"}</p>}
    {inbox.received.map((item) => <ContactRequestRow key={`received:${item.request_id}`} item={item} tone="received" onOpen={onOpen}><div className="contact-request-actions"><button className="secondary-button" disabled={Boolean(busy)} onClick={() => void onRespond(item, false)}>拒绝</button><button className="primary-button" disabled={Boolean(busy) || contactItemExpired(item)} onClick={() => void onRespond(item, true)}>接受并互相授权</button></div></ContactRequestRow>)}
    {inbox.sent.map((item) => <ContactRequestRow key={`sent:${item.request_id}`} item={item} tone="sent" onOpen={onOpen}><small className="contact-request-status">{contactItemExpired(item) ? "已过期，可打开详情重新申请" : "等待好友处理"}</small></ContactRequestRow>)}
    {inbox.reapproval.map((item) => <ContactRequestRow key={`reapproval:${item.friend_id}`} item={item} tone="reapproval" onOpen={onOpen}><button className="secondary-button" disabled={Boolean(busy)} onClick={() => void onRequest(item)}><RotateCcw size={15} />重新申请</button></ContactRequestRow>)}
  </div>
}

function ContactRequestRow({ item, tone, onOpen, children }: { item: ContactInboxItem; tone: "received" | "sent" | "reapproval"; onOpen: (item: ContactInboxItem) => void; children: ReactNode }) {
  const status = tone === "received" ? "收到申请" : tone === "sent" ? "已发出申请" : item.status === "revoked" ? "授权已撤销" : "联系方式已变化，需要重新授权"
  return <article className={`contact-request-row ${tone}`}><button className="contact-request-main" onClick={() => onOpen(item)}><span className="friend-avatar">{item.username?.slice(0, 2) || "??"}</span><span><b>{code(item.username)}</b><small>{status} · {channelLabel(item.channels)} · {formatContactTime(item.requested_at)}</small>{item.expires_at && tone !== "reapproval" && <small>{contactItemExpired(item) ? "已过期" : `有效期至 ${formatContactTime(item.expires_at)}`}</small>}</span></button>{children}</article>
}

export function ExchangeCard({ exchange, me, busy, onRequest, onRespond }: { exchange: ContactExchange; me: string; busy?: boolean; onRequest: () => Promise<void>; onRespond: (accept: boolean) => Promise<void> }) {
  const contact = disclosedFriendContact(exchange)
  if (["none", "declined", "expired", "cancelled", "stale", "revoked"].includes(exchange.status)) {
    const copy = exchange.status === "stale" ? "任一方的联系方式已变化，旧授权已整体失效，双方均不能再读取原联系方式。" : exchange.status === "revoked" ? "这次授权已被安全撤销；重新成为好友也不会恢复旧授权。" : exchange.status === "expired" ? "这条申请已过期，需要重新发起。" : exchange.status === "declined" ? "上一条申请已被拒绝，你可以在合适的时候重新申请。" : "向好友发起双向授权；请求本身不会包含 QQ 或微信正文。"
    return <div className={`exchange-state ${exchange.status}`}><AlertTriangle size={28} /><b>{exchange.status === "stale" ? "联系方式已变化，需要重新授权" : exchange.status === "revoked" ? "授权已撤销" : "尚未完成授权"}</b><p>{copy}</p><button className="primary-button" disabled={busy} onClick={() => void onRequest()}><Plus size={17} />重新申请交换已保存的联系方式</button></div>
  }
  if (exchange.status === "pending") {
    const canRespond = exchange.recipient_id === me
    return <div className="exchange-state pending"><b>交换请求等待处理</b><p>发起方已保存：{channelLabel(exchange.channels)}</p>{canRespond ? <><small>你只需保存 QQ 或微信中的任意一种；接受后，双方各自授权自己已保存的联系方式。</small><div className="modal-actions"><button className="secondary-button" disabled={busy} onClick={() => void onRespond(false)}>拒绝</button><button className="primary-button" disabled={busy} onClick={() => void onRespond(true)}>接受并互相授权</button></div></> : <small>好友接受后，双方才能看到彼此实际提供的联系方式。</small>}</div>
  }
  return <div className="exchange-state accepted"><b><Check size={18} />已互相授权</b><p>好友 QQ：<strong>{contact?.qq || "未提供"}</strong></p><p>好友微信：<strong>{contact?.wechat || "未提供"}</strong></p><small>任一方修改或清空任何联系方式后，本页会停止返回全部联系方式，直至重新授权。</small></div>
}
