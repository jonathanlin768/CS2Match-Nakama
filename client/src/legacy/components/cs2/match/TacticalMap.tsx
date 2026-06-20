import { useRef } from "react";

interface MapPlayer {
  id: string;
  name: string;
  x: number;
  y: number;
  team: "t" | "ct";
  alive: boolean;
}

interface Props {
  mapName: string;
  players: MapPlayer[];
  currentTime: string;
  totalTime: string;
}

const MAP_IMAGE = "/assets/images/maps/de_dust2_radar_trans.webp";

export function TacticalMap({ players, currentTime }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);

  return (
    <div ref={containerRef} className="relative w-full h-full bg-card border border-border rounded-lg overflow-hidden">
      {/* Map background */}
      <img
        src={MAP_IMAGE}
        alt="Dust2"
        className="absolute inset-0 w-full h-full object-contain opacity-80"
        draggable={false}
      />

      {/* Overlay players */}
      {players.map((p) => (
        <div
          key={p.id}
          className={`absolute w-3 h-3 rounded-full border-2 border-white shadow-sm transition-all ${
            p.alive ? (p.team === "t" ? "bg-orange-500" : "bg-blue-500") : "bg-gray-500 opacity-50"
          }`}
          style={{
            left: `${p.x}%`,
            top: `${p.y}%`,
            transform: "translate(-50%, -50%)",
          }}
          title={p.name}
        >
          {/* Name label */}
          <span
            className={`absolute top-4 left-1/2 -translate-x-1/2 text-[10px] font-bold whitespace-nowrap ${
              p.team === "t" ? "text-orange-400" : "text-blue-400"
            }`}
          >
            {p.name}
          </span>
        </div>
      ))}

      {/* Time badge */}
      <div className="absolute top-2 left-1/2 -translate-x-1/2 px-2 py-1 bg-black/70 text-white text-xs font-mono rounded">
        {currentTime}
      </div>
    </div>
  );
}
