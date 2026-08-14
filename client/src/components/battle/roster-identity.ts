import { createElement, Fragment, type ReactNode } from "react"
import type { BattlePlayer } from "./data/battle"

type IdentifiedBattlePlayer = Pick<BattlePlayer, "id" | "instanceId" | "configPlayerId">

export function battleRosterRows<T extends IdentifiedBattlePlayer>(players: readonly T[]) {
  return players.map((player, index) => ({
    player,
    key: player.instanceId ?? `${player.id}:${index}`,
    playerId: player.instanceId,
    configPlayerId: player.configPlayerId,
  }))
}

export function BattleRosterIdentityRows<T extends IdentifiedBattlePlayer>({
  players,
  className,
  renderPlayer,
}: {
  players: readonly T[]
  className?: string
  renderPlayer: (player: T) => ReactNode
}) {
  return createElement(
    Fragment,
    null,
    ...battleRosterRows(players).map(({ player, key, playerId, configPlayerId }) =>
      createElement(
        "div",
        { key, "data-player-id": playerId, "data-config-player-id": configPlayerId, className },
        renderPlayer(player),
      ),
    ),
  )
}
