import { useMemo } from "react";
import { useSimStream } from "@/hooks/useSimStream";
import { MatchScoreBar } from "@/components/cs2/match/MatchScoreBar";
import { TeamRoster } from "@/components/cs2/match/TeamRoster";
import { TacticalMap } from "@/components/cs2/match/TacticalMap";
import { RoundEvents } from "@/components/cs2/match/RoundEvents";
import { PlayerStats } from "@/components/cs2/match/PlayerStats";
import { RoundTactics } from "@/components/cs2/match/RoundTactics";

const team1 = {
  name: "Terrorists",
  logo: "🔴",
  score: 0,
};

const team2 = {
  name: "Counter-Terrorists",
  logo: "🟡",
  score: 0,
};

// Radar canvas dimensions (must match server map)
const MAP_W = 1024;
const MAP_H = 984;

export default function MatchPage() {
  const {
    connected,
    tickMsg,
    killEvents,
    c4Events,
    roundEnd,
    players,
  } = useSimStream();

  // Note: useSimStream is a stub — no WebSocket, no auto-start
  // All derived data naturally resolves to empty/placeholder values

  // Convert pixel coords to percentage for TacticalMap
  const mapPlayers = useMemo(() => {
    return players.map((p) => ({
      id: p.id,
      name: p.name,
      x: (p.x / MAP_W) * 100,
      y: (p.y / MAP_H) * 100,
      team: p.team.toLowerCase() as "t" | "ct",
      alive: p.alive,
    }));
  }, [players]);

  // Death markers (persistent for the round)
  const deathMarkers = useMemo(() => {
    return killEvents.map((k) => ({
      x: (k.deathX / MAP_W) * 100,
      y: (k.deathY / MAP_H) * 100,
    }));
  }, [killEvents]);

  // C4 overlay
  const c4State = tickMsg?.c4State;
  const c4Pos = tickMsg?.c4Pos
    ? {
        x: (tickMsg.c4Pos.x / MAP_W) * 100,
        y: (tickMsg.c4Pos.y / MAP_H) * 100,
      }
    : undefined;

  // Timer text
  const remainingTicks = tickMsg ? tickMsg.maxTick - tickMsg.tick : 115;
  const minutes = Math.floor(remainingTicks / 60);
  const seconds = remainingTicks % 60;
  const roundTime = `${minutes}:${String(seconds).padStart(2, "0")}`;

  // Convert kill events to RoundEvents format
  const roundEventItems = useMemo(() => {
    return killEvents.map((k) => ({
      time: `${k.tick}s`,
      type: "kill" as const,
      player: k.attackerName || "?",
      playerTeam: (players.find((p) => p.id === k.attackerId)?.team?.toLowerCase() || "t") as "t" | "ct",
      target: k.victimName,
      targetTeam: (players.find((p) => p.id === k.victimId)?.team?.toLowerCase() || "ct") as "t" | "ct",
      weapon: k.weapon,
      description: k.headshot ? "爆头击杀了" : "击杀了",
    }));
  }, [killEvents, players]);

  // Convert C4 events to RoundEvents format
  const c4EventItems = useMemo(() => {
    return c4Events.map((e) => {
      const player = players.find((p) => p.id === e.playerId);
      let description = "";
      switch (e.type) {
        case "plant_started":
          description = "开始安装 C4";
          break;
        case "plant_finished":
          description = `在 ${e.site || "?"} 点安装 C4`;
          break;
        case "defuse_started":
          description = "开始拆除 C4";
          break;
        case "defuse_finished":
          description = "成功拆除 C4";
          break;
        case "exploded":
          description = "C4 爆炸";
          break;
      }
      return {
        time: `${e.tick}s`,
        type: "bomb" as const,
        player: player?.name || "?",
        playerTeam: (player?.team?.toLowerCase() || "t") as "t" | "ct",
        target: "",
        targetTeam: "t" as "t" | "ct",
        weapon: "💣",
        description,
      };
    });
  }, [c4Events, players]);

  const allEvents = useMemo(() => {
    return [...c4EventItems, ...roundEventItems];
  }, [c4EventItems, roundEventItems]);

  // Roster data (adapt SimPlayerState to TeamRoster format)
  const tRoster = useMemo(() => {
    return players
      .filter((p) => p.team === "T")
      .map((p) => ({
        id: p.id,
        name: p.name,
        avatar: "👤",
        weapons: ["🔫", "🔪"],
        money: 0,
        health: p.hp,
        kills: 0,
        deaths: p.alive ? 0 : 1,
      }));
  }, [players]);

  const ctRoster = useMemo(() => {
    return players
      .filter((p) => p.team === "CT")
      .map((p) => ({
        id: p.id,
        name: p.name,
        avatar: "👮",
        weapons: ["🔫", "🔪"],
        money: 0,
        health: p.hp,
        kills: 0,
        deaths: p.alive ? 0 : 1,
      }));
  }, [players]);

  // Tactics display
  const tacticItems = useMemo(() => {
    const items: { time: string; type: "tactic" | "kill" | "smoke"; title: string; team: "t" | "ct"; description: string }[] = [];
    if (tickMsg) {
      items.push({
        time: "R1",
        type: "tactic",
        title: "T 方战术",
        team: "t",
        description: tickMsg.players.length > 0 ? "RushB" : "等待中",
      });
    }
    return items;
  }, [tickMsg]);

  return (
    <div className="flex-1 flex flex-col">
      {/* Connection status indicator */}
      <div className="px-4 py-1 text-xs flex items-center gap-2 bg-card border-b border-border">
        <span
          className={`inline-block w-2 h-2 rounded-full ${
            connected ? "bg-green-500" : "bg-red-500"
          }`}
        />
        <span className="text-muted-foreground">
          {connected ? "已连接模拟服务器" : "未连接"}
        </span>
        {roundEnd && (
          <span className="ml-auto font-semibold text-primary">
            回合结束: {roundEnd.winner === "T" ? "T 方胜利" : "CT 方胜利"} ({roundEnd.reason})
          </span>
        )}
      </div>

      <MatchScoreBar
        team1={team1}
        team2={team2}
        roundTime={roundTime}
        currentRound={1}
        totalRounds={1}
      />

      <div className="flex-1 p-4 lg:p-6">
        <div className="max-w-[1400px] mx-auto flex flex-col gap-4 lg:gap-6">
          {/* Main Content Area */}
          <div className="flex flex-col lg:flex-row gap-4 lg:gap-6">
            {/* Left Roster */}
            <div className="w-full lg:w-56 xl:w-64 order-2 lg:order-1">
              <TeamRoster
                teamName="Terrorists"
                teamLogo="🔴"
                teamMoney={0}
                players={tRoster}
                side="left"
                teamColor="orange"
              />
            </div>

            {/* Center: Tactical Map */}
            <div className="flex-1 order-1 lg:order-2 min-h-[300px] lg:min-h-[400px]">
              <TacticalMap
                mapName="Dust2"
                players={mapPlayers}
                currentTime={roundTime}
                totalTime="1:55"
              />
              {/* Overlay C4 / death markers if TacticalMap doesn't support them natively */}
              <div className="relative w-full h-full -mt-full pointer-events-none">
                {c4State === "planted" && c4Pos && (
                  <div
                    className="absolute w-4 h-4 bg-orange-500 rounded-full animate-pulse border-2 border-white"
                    style={{ left: `${c4Pos.x}%`, top: `${c4Pos.y}%`, transform: "translate(-50%, -50%)" }}
                    title="C4 Planted"
                  />
                )}
                {deathMarkers.map((m, i) => (
                  <div
                    key={i}
                    className="absolute text-red-500 font-bold text-lg"
                    style={{ left: `${m.x}%`, top: `${m.y}%`, transform: "translate(-50%, -50%)" }}
                  >
                    ✕
                  </div>
                ))}
              </div>
            </div>

            {/* Right Roster */}
            <div className="w-full lg:w-56 xl:w-64 order-3">
              <TeamRoster
                teamName="Counter-Terrorists"
                teamLogo="🟡"
                teamMoney={0}
                players={ctRoster}
                side="right"
                teamColor="blue"
              />
            </div>
          </div>

          {/* Bottom Section */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 lg:gap-6">
            {/* Round Events */}
            <div className="lg:col-span-3 min-h-[200px] lg:min-h-[280px]">
              <RoundEvents events={allEvents} />
            </div>

            {/* Player Stats */}
            <div className="lg:col-span-6">
              <PlayerStats
                team1Stats={tRoster.map((p) => ({
                  name: p.name,
                  team: "t" as const,
                  kills: 0,
                  assists: 0,
                  deaths: p.deaths,
                  kd: p.deaths > 0 ? 0 : 0,
                  adr: 0,
                  hsPercent: 0,
                  kast: "-",
                  rating: 1.0,
                }))}
                team2Stats={ctRoster.map((p) => ({
                  name: p.name,
                  team: "ct" as const,
                  kills: 0,
                  assists: 0,
                  deaths: p.deaths,
                  kd: p.deaths > 0 ? 0 : 0,
                  adr: 0,
                  hsPercent: 0,
                  kast: "-",
                  rating: 1.0,
                }))}
                team1Name="Terrorists"
                team2Name="Counter-Terrorists"
                team1Logo="🔴"
                team2Logo="🟡"
              />
            </div>

            {/* Round Tactics */}
            <div className="lg:col-span-3 min-h-[200px] lg:min-h-[280px]">
              <RoundTactics tactics={tacticItems} currentRound={1} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
