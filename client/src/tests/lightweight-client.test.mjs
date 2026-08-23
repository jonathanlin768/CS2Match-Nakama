import assert from "node:assert/strict"
import test from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { countUnreadCards, parseContactCard, shouldRefreshContactInbox } from "../social/contact-card.ts"
import { groupFriends } from "../social/friend-groups.ts"
import { toggleTutorialPlayer, tutorialSelectionCost, tutorialSelectionReady } from "../game/tutorial-selection.ts"
import { assetUrl, croppedImageStyle, validAvatarCrop } from "../components/player/player-visual.ts"
import { rpcErrorMessage } from "../api/rpc-error.ts"
import { BattleRosterIdentityRows, battleRosterRows } from "../components/battle/roster-identity.ts"
import { decodeSocialRpcPayload, SocialApiError, toSocialApiError } from "../api/social-error.ts"
import { contactChannels, contactItemExpired, contactProfileChanged, disclosedFriendContact, mergeContactInboxPages } from "../social/contact-inbox.ts"

const tutorial = {
  budget: 15, rosterSize: 5,
  tier1PlayerIds: ["p1"], tier2PlayerIds: ["p2"], tier3PlayerIds: ["p3"], tier4PlayerIds: ["p4"], tier5PlayerIds: ["p5", "p6", "p7"],
}

test("tutorial selection enforces pool, roster and budget boundaries", () => {
  let selected = []
  for (const id of ["p1", "p2", "p3", "p4", "p5"]) selected = toggleTutorialPlayer(tutorial, selected, id).selected
  assert.equal(tutorialSelectionCost(tutorial, selected), 15)
  assert.equal(tutorialSelectionReady(tutorial, selected), true)
  assert.equal(toggleTutorialPlayer(tutorial, selected, "p6").error, "full")
  assert.equal(toggleTutorialPlayer(tutorial, [], "missing").error, "unknown")
  assert.equal(toggleTutorialPlayer(tutorial, ["p5", "p6", "p1"], "p7").error, "budget")
  assert.deepEqual(toggleTutorialPlayer(tutorial, selected, "p3").selected, ["p1", "p2", "p4", "p5"])
})

test("RPC response errors expose the Nakama message", async () => {
  const response = new Response(JSON.stringify({ code: 13, error: {}, message: "INVALID_LINEUP: duplicate player id: p1" }), { status: 500, statusText: "Internal Server Error" })
  assert.equal(await rpcErrorMessage(response), "INVALID_LINEUP: duplicate player id: p1")
  assert.equal(await rpcErrorMessage(new Error("offline")), "offline")
})

test("player visual helpers preserve normalized 2:3 to 5:7 crops", () => {
  const crop = { x: 0.2, y: 0.08, width: 0.6, height: 0.56 }
  assert.equal(validAvatarCrop(crop), true)
  assert.equal(validAvatarCrop({ ...crop, x: 0.8 }), false)
  assert.equal(validAvatarCrop({ ...crop, height: 0.2 }), false)
  assert.deepEqual(croppedImageStyle(crop), {
    width: "166.66666666666669%",
    maxWidth: "none",
    height: "auto",
    left: "-33.333333333333336%",
    top: "-14.285714285714285%",
  })
  assert.equal(assetUrl("player-cards/niko2.png"), "/player-cards/niko2.png")
  assert.equal(assetUrl("https://example.com/player.png"), "https://example.com/player.png")
  assert.equal(assetUrl(), "/images/star-player.png")
})

test("battle roster renders shared config players as two distinct instance rows", () => {
  const cardImage = "player-cards/zywoo.png"
  const rows = battleRosterRows([
    { id: "ZyWOo", instanceId: "tutorial_players/player_zywoo", configPlayerId: "player_zywoo", cardImage },
    { id: "ZyWOo", instanceId: "team_vitality/player_zywoo", configPlayerId: "player_zywoo", cardImage },
  ])

  assert.equal(rows.length, 2)
  assert.deepEqual(rows.map((row) => row.key), ["tutorial_players/player_zywoo", "team_vitality/player_zywoo"])
  assert.deepEqual(rows.map((row) => row.playerId), ["tutorial_players/player_zywoo", "team_vitality/player_zywoo"])
  assert.deepEqual(rows.map((row) => row.configPlayerId), ["player_zywoo", "player_zywoo"])
  assert.deepEqual(rows.map((row) => row.player.cardImage), [cardImage, cardImage])

  const markup = renderToStaticMarkup(createElement(BattleRosterIdentityRows, {
    players: rows.map((row) => row.player),
    renderPlayer: (player) => createElement("span", { "data-card-image": player.cardImage }, player.id),
  }))
  assert.equal((markup.match(/data-player-id=/g) ?? []).length, 2)
  assert.match(markup, /data-player-id="tutorial_players\/player_zywoo"/)
  assert.match(markup, /data-player-id="team_vitality\/player_zywoo"/)
  assert.equal((markup.match(/data-config-player-id="player_zywoo"/g) ?? []).length, 2)
  assert.equal((markup.match(/data-card-image="player-cards\/zywoo.png"/g) ?? []).length, 2)
})

test("contact card parser accepts only supported server card shapes", () => {
  assert.deepEqual(parseContactCard({ message_id: "m1", sender_id: "friend", content: JSON.stringify({ type: "contact_exchange", request_id: "r1", action: "requested", version: 2 }) }), { messageId: "m1", requestId: "r1", action: "requested", version: 2, createTime: undefined, senderId: "friend" })
  assert.equal(parseContactCard({ content: JSON.stringify({ type: "text", text: "hello" }) }), null)
  assert.equal(parseContactCard({ content: "not json" }), null)
  assert.equal(parseContactCard({ content: { type: "contact_exchange", request_id: "r2", action: "unknown", version: 1 } }), null)
  assert.equal(parseContactCard({ content: { type: "contact_exchange", request_id: "r3", action: "stale", version: 3 } })?.action, "stale")
  assert.equal(parseContactCard({ content: { type: "contact_exchange", request_id: "r4", action: "revoked", version: 4 } })?.action, "revoked")
})

test("contact card unread count includes only incoming cards newer than the read cursor", () => {
  const cards = [
    { messageId: "old", requestId: "r", action: "requested", version: 1, createTime: "2026-08-09T09:00:00Z", senderId: "friend" },
    { messageId: "mine", requestId: "r", action: "accepted", version: 2, createTime: "2026-08-09T10:00:00Z", senderId: "me" },
    { messageId: "new", requestId: "r", action: "accepted", version: 3, createTime: "2026-08-09T11:00:00Z", senderId: "friend" },
  ]
  assert.equal(countUnreadCards(cards, "me", "2026-08-09T09:30:00Z"), 1)
})

test("friend groups preserve accepted, outgoing and incoming Nakama states", () => {
  const grouped = groupFriends([{ id: "a", state: 0 }, { id: "b", state: 1 }, { id: "c", state: 2 }])
  assert.deepEqual(grouped.accepted.map((item) => item.id), ["a"])
  assert.deepEqual(grouped.outgoing.map((item) => item.id), ["b"])
  assert.deepEqual(grouped.incoming.map((item) => item.id), ["c"])
})

test("social RPC payload decoder supports Nakama string and object payloads", () => {
  assert.deepEqual(decodeSocialRpcPayload('{"status":"pending","version":2}'), { status: "pending", version: 2 })
  assert.deepEqual(decodeSocialRpcPayload({ status: "accepted", version: 3 }), { status: "accepted", version: 3 })
})

test("social errors map Response, SDK and network failures without leaking payloads", async () => {
  const response = new Response(JSON.stringify({ code: 13, message: "PROFILE_INCOMPLETE: qq=123456789 wechat=secret_user" }), { status: 400 })
  const mappedResponse = await toSocialApiError(response)
  assert.equal(mappedResponse.code, "PROFILE_INCOMPLETE")
  assert.equal(mappedResponse.message, "请先保存至少一种联系方式；QQ 或微信任选一项即可，不需要两项都填写。")
  assert.doesNotMatch(mappedResponse.message, /123456789|secret_user/)

  const sdk = await toSocialApiError({ code: "CONFLICT", message: "raw private contact 99887766" })
  assert.equal(sdk.message, "状态已发生变化，请刷新后重试。")
  const network = await toSocialApiError(new Error("offline"))
  assert.equal(network.message, "网络连接失败，请检查连接后重试。")
  const unknown = await toSocialApiError({ message: "qq=99887766" })
  assert.equal(unknown.message, "联系方式服务暂时不可用，请稍后重试。")
  assert.equal(await toSocialApiError(new SocialApiError("NOT_FRIENDS")) instanceof SocialApiError, true)
})

test("authoritative inbox pages merge independently of optional DM cards", () => {
  const merged = mergeContactInboxPages([
    { received: [{ friend_id: "b", username: "BBBBBBBB", status: "pending" }], sent: [], reapproval: [], incoming_pending_count: 1, cursor: "next" },
    { received: [], sent: [{ friend_id: "c", username: "CCCCCCCC", status: "pending" }], reapproval: [{ friend_id: "d", username: "DDDDDDDD", status: "stale" }], incoming_pending_count: 0 },
  ])
  assert.deepEqual(merged.received.map((item) => item.friend_id), ["b"])
  assert.deepEqual(merged.sent.map((item) => item.friend_id), ["c"])
  assert.deepEqual(merged.reapproval.map((item) => item.status), ["stale"])
  assert.equal(merged.incoming_pending_count, 1)
  assert.equal("qq" in merged || "wechat" in merged, false)
})

test("inbox refresh policy recovers on login, visibility and reconnect and ignores invalid cards", () => {
  for (const reason of ["login", "explicit", "visible", "reconnect", "notification"]) assert.equal(shouldRefreshContactInbox(reason), true)
  const friendOutsideCurrentDetail = { sender_id: "different-friend", content: { type: "contact_exchange", request_id: "request-other", action: "requested", version: 1 } }
  assert.equal(shouldRefreshContactInbox("card", friendOutsideCurrentDetail), true)
  assert.equal(shouldRefreshContactInbox("card", { content: { type: "contact_exchange", action: "requested" } }), false)
  assert.equal(shouldRefreshContactInbox("card_history_failure"), false)
})

test("profile changes invalidate the whole exchange and only accepted state discloses contact", () => {
  const profile = { qq: "123456", wechat: "wx_user", revision: 2 }
  assert.equal(contactProfileChanged(profile, "123456", "wx_user"), false)
  assert.equal(contactProfileChanged(profile, "123456", ""), true)
  assert.deepEqual(contactChannels({ qq: "123456" }), ["qq"])
  assert.deepEqual(contactChannels({ wechat: "wx_user" }), ["wechat"])
  assert.deepEqual(contactChannels({ qq: "123456", wechat: "wx_user" }), ["qq", "wechat"])
  assert.equal(contactItemExpired({ friend_id: "b", username: "B", status: "pending", expires_at: 50 }, 51), true)
  assert.deepEqual(disclosedFriendContact({ status: "accepted", friend_contact: { qq: "123456", wechat: "wx_user" } }), { qq: "123456", wechat: "wx_user" })
  assert.equal(disclosedFriendContact({ status: "stale", friend_contact: { qq: "must-not-show", wechat: "must-not-show" } }), null)
  assert.equal(disclosedFriendContact({ status: "revoked", friend_contact: { qq: "must-not-show" } }), null)
})
