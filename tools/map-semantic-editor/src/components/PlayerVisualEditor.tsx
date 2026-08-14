import { ImagePlus, RotateCcw } from 'lucide-react'
import { useMemo, useRef, useState } from 'react'
import ReactCrop, { centerCrop, makeAspectCrop, type PercentCrop } from 'react-image-crop'
import 'react-image-crop/dist/ReactCrop.css'
import { toNumberValue } from '../lib/configValues'
import { configAssetUrl } from '../lib/api'
import { configFieldLabel } from '../lib/configLabels'
import type { LubanCellValue, LubanRow } from '../lib/luban'
import { useConfigStore } from '../store/configStore'

const aspect = 5 / 7

export function PlayerVisualEditor({ row, onChange }: { row: LubanRow; onChange: (values: Record<string, LubanCellValue>) => void }) {
  const inputRef = useRef<HTMLInputElement>(null)
  const imageRef = useRef<HTMLImageElement>(null)
  const resetAfterLoad = useRef(false)
  const uploadImage = useConfigStore((state) => state.uploadImage)
  const [missing, setMissing] = useState(false)
  const cardImage = String(row.cardImage ?? '')
  const portrait = String(row.portrait ?? '')
  const source = cardImage || portrait
  const crop = useMemo<PercentCrop>(() => ({
    unit: '%',
    x: toNumberValue(row.avatarCropX) * 100,
    y: toNumberValue(row.avatarCropY) * 100,
    width: toNumberValue(row.avatarCropWidth) * 100,
    height: toNumberValue(row.avatarCropHeight) * 100,
  }), [row.avatarCropHeight, row.avatarCropWidth, row.avatarCropX, row.avatarCropY])

  async function selectFile(file?: File) {
    if (!file) return
    const assetPath = await uploadImage('player-card', file)
    if (assetPath) {
      resetAfterLoad.current = true
      setMissing(false)
      onChange({ cardImage: assetPath, avatarCropX: 0, avatarCropY: 0, avatarCropWidth: 0, avatarCropHeight: 0 })
    }
    if (inputRef.current) inputRef.current.value = ''
  }

  function resetCrop(image = imageRef.current) {
    if (!image) return
    const next = centerCrop(makeAspectCrop({ unit: '%', width: 100 }, aspect, image.naturalWidth, image.naturalHeight), image.naturalWidth, image.naturalHeight)
    writeCrop(next)
  }

  function writeCrop(next: PercentCrop) {
    onChange({
      avatarCropX: round(next.x / 100),
      avatarCropY: round(next.y / 100),
      avatarCropWidth: round(next.width / 100),
      avatarCropHeight: round(next.height / 100),
    })
  }

  const validCrop = crop.width > 0 && crop.height > 0
  const previewStyle = validCrop && cardImage ? cropBackgroundStyle(configAssetUrl(cardImage), crop) : undefined

  return (
    <div className="playerVisualEditor">
      <div className="visualSourceColumn">
        <h4>完整卡面</h4>
        <div className="fullCardPreview">{cardImage && !missing ? <img src={configAssetUrl(cardImage)} alt="完整选手卡面" onError={() => setMissing(true)} /> : <span>{missing ? '卡面图片缺失' : '尚未配置完整卡面'}</span>}</div>
        <input ref={inputRef} className="visuallyHidden" type="file" accept="image/png,image/jpeg,image/webp" onChange={(event) => void selectFile(event.target.files?.[0])} />
        <button type="button" className="commandButton" onClick={() => inputRef.current?.click()}><ImagePlus size={16} />上传卡面</button>
        <label className="configField"><span>{configFieldLabel('cardImage')}</span><input value={cardImage} onChange={(event) => { setMissing(false); onChange({ cardImage: event.target.value }) }} /></label>
      </div>

      <div className="visualCropColumn">
        <div className="visualColumnHeading"><h4>头像裁切</h4><button type="button" className="iconTextButton" disabled={!source || missing} onClick={() => resetCrop()}><RotateCcw size={15} />重置裁切</button></div>
        <div className="cropCanvas">
          {source && !missing ? <ReactCrop crop={crop} aspect={aspect} keepSelection onChange={(_, percentCrop) => writeCrop(percentCrop)}>
            <img ref={imageRef} src={configAssetUrl(source)} alt="头像裁切源图" onError={() => setMissing(true)} onLoad={(event) => { setMissing(false); if (resetAfterLoad.current || !validCrop) { resetAfterLoad.current = false; resetCrop(event.currentTarget) } }} />
          </ReactCrop> : <span>{missing ? '图片无法读取' : '上传卡面后可设置裁切'}</span>}
        </div>
      </div>

      <div className="visualResultColumn">
        <h4>战斗头像预览</h4>
        <div className="avatarCropPreview" style={previewStyle}>{!previewStyle ? <span>5:7</span> : null}</div>
        <dl className="cropValues">
          <div><dt>{configFieldLabel('avatarCropX')}</dt><dd>{(crop.x / 100).toFixed(4)}</dd></div>
          <div><dt>{configFieldLabel('avatarCropY')}</dt><dd>{(crop.y / 100).toFixed(4)}</dd></div>
          <div><dt>{configFieldLabel('avatarCropWidth')}</dt><dd>{(crop.width / 100).toFixed(4)}</dd></div>
          <div><dt>{configFieldLabel('avatarCropHeight')}</dt><dd>{(crop.height / 100).toFixed(4)}</dd></div>
        </dl>
      </div>
    </div>
  )
}

function cropBackgroundStyle(source: string, crop: PercentCrop) {
  const x = crop.width >= 100 ? 50 : (crop.x / (100 - crop.width)) * 100
  const y = crop.height >= 100 ? 50 : (crop.y / (100 - crop.height)) * 100
  return {
    backgroundImage: `url("${source.replace(/"/g, '%22')}")`,
    backgroundSize: `${10000 / crop.width}% auto`,
    backgroundPosition: `${x}% ${y}%`,
  }
}

function round(value: number): number {
  return Math.round(value * 1000000) / 1000000
}
