import { assetUrl, DEFAULT_PLAYER_IMAGE } from "./player-visual"

export default function PlayerFullCard({ cardImage, portrait, alt, className = "" }: { cardImage?: string; portrait?: string; alt: string; className?: string }) {
  const primary = cardImage || portrait

  return <div className={`player-full-card ${className}`}>
    <img
      src={assetUrl(primary)}
      alt={alt}
      data-stage={cardImage ? "card" : portrait ? "portrait" : "default"}
      onError={(event) => {
        const image = event.currentTarget
        if (image.dataset.stage === "card" && portrait) {
          image.dataset.stage = "portrait"
          image.src = assetUrl(portrait)
          return
        }
        if (image.dataset.stage !== "default") {
          image.dataset.stage = "default"
          image.src = DEFAULT_PLAYER_IMAGE
          return
        }
        image.onerror = null
      }}
    />
  </div>
}
