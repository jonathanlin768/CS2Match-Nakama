import assert from "node:assert/strict"
import test from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { countUnreadCards, parseContactCard } from "../social/contact-card.ts"
import { groupFriends } from "../social/friend-groups.ts"
import { toggleTutorialPlayer, tutorialSelectionCost, tutorialSelectionReady } from "../game/tutorial-selection.ts"
import { assetUrl, croppedImageStyle, validAvatarCrop } from "../components/player/player-visual.ts"
import { rpcErrorMessage } from "../api/rpc-error.ts"
import { BattleRosterIdentityRows, battleRosterRows } from "../components/battle/roster-identity.ts"

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
