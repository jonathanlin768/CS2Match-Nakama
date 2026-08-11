export interface ImageCrop {
  x: number
  y: number
  width: number
  height: number
}

export const DEFAULT_PLAYER_IMAGE = "/images/star-player.png"

export function assetUrl(source?: string): string {
  if (!source) return DEFAULT_PLAYER_IMAGE
  if (source.startsWith("/") || source.startsWith("http://") || source.startsWith("https://") || source.startsWith("data:")) return source
  return `/${source.replace(/^\.\//, "")}`
}

export function validAvatarCrop(crop?: ImageCrop): crop is ImageCrop {
  if (!crop) return false
  const values = [crop.x, crop.y, crop.width, crop.height]
  if (values.some((value) => !Number.isFinite(value))) return false
  if (crop.x < 0 || crop.y < 0 || crop.width <= 0 || crop.height <= 0 || crop.x + crop.width > 1.000001 || crop.y + crop.height > 1.000001) return false
  return Math.abs((crop.width * 2) / (crop.height * 3) - 5 / 7) <= 0.005
}

export function croppedImageStyle(crop: ImageCrop) {
  return {
    width: `${100 / crop.width}%`,
    maxWidth: "none",
    height: "auto",
    left: `${(-crop.x / crop.width) * 100}%`,
    top: `${(-crop.y / crop.height) * 100}%`,
  }
}
