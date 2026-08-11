import { assetUrl, croppedImageStyle, DEFAULT_PLAYER_IMAGE, validAvatarCrop, type ImageCrop } from "./player-visual"

export default function PlayerAvatarCrop({ cardImage, portrait, crop, alt, className = "" }: { cardImage?: string; portrait?: string; crop?: ImageCrop; alt: string; className?: string }) {
  const useCrop = Boolean(cardImage && validAvatarCrop(crop))
  const source = useCrop ? cardImage : portrait

  return <div className={`player-avatar-crop ${className}`}>
    <img
      src={assetUrl(source)}
      alt={alt}
      data-stage={useCrop ? "card" : portrait ? "portrait" : "default"}
      className={useCrop ? "player-avatar-cropped-source" : "player-avatar-fallback"}
      style={useCrop ? croppedImageStyle(crop!) : undefined}
      onError={(event) => {
        const image = event.currentTarget
        if (image.dataset.stage === "card" && portrait) {
          image.dataset.stage = "portrait"
          image.src = assetUrl(portrait)
          image.className = "player-avatar-fallback"
          image.removeAttribute("style")
          return
        }
        if (image.dataset.stage !== "default") {
          image.dataset.stage = "default"
          image.src = DEFAULT_PLAYER_IMAGE
          image.className = "player-avatar-fallback"
          image.removeAttribute("style")
          return
        }
        image.onerror = null
      }}
    />
  </div>
}
