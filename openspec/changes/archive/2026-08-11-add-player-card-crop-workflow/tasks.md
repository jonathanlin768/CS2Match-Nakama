## 1. Luban Player 数据结构

- [x] 1.1 在 `configs/Datas/#Player.xlsx` 按现有 Luban 表头规范增加 `cardImage`、`avatarCropX`、`avatarCropY`、`avatarCropWidth`、`avatarCropHeight` 字段，并为已有 Player 行设置兼容旧数据的空路径和零值默认值
- [x] 1.2 运行 `scripts/gen-config.ps1`，更新 Server Go、Client TypeScript 和 JSON 生成产物，并核对新增字段的名称、类型及值在三端一致
- [x] 1.3 扩展配置加载测试，覆盖包含新字段的 Player、完全缺少卡面的旧 Player，以及零值裁切数据
- [x] 1.4 将用户提供的 23 张卡面复制到 `client/public/player-cards/`，为 18 名可可靠匹配的现有 Player 写入默认卡面和合法 `5:7` 裁切，并确认 `chopper`、`kyxsan` 继续使用旧头像回退

## 2. 配置服务与资源写入

- [x] 2.1 为配置服务增加 `player-card` 图片类型和 `client/public/player-cards/` 受控目录，支持 PNG、JPEG、WebP 的复制、重名确认和项目相对路径返回
- [x] 2.2 扩展上传 API、前端 API 类型和配置 store，使 Player 页面能够上传完整卡面且不改变现有 `portrait`、Team Logo 上传行为
- [x] 2.3 增加配置服务测试，覆盖卡面成功写入、格式/大小限制、目录穿越拒绝、重名拒绝与显式覆盖

## 3. Player 裁切编辑体验

- [x] 3.1 在配置工具中加入并锁定 `react-image-crop` 依赖及样式，封装固定 `5:7` 的 percent crop 与 `0..1` 归一化参数转换
- [x] 3.2 实现居中最大 `5:7` 默认裁切、拖动、缩放、重置和实时预览，并保证计算基于图片自然尺寸而非页面显示尺寸
- [x] 3.3 按设计线框扩展 Player 专属页的“选手视觉资源”区域，提供上传卡面、完整 `2:3` 预览、裁切画布、`5:7` 结果预览和参数只读显示
- [x] 3.4 将裁切参数接入 Player 行编辑、脏状态、保存、重新读取与写入 Luban 流程；更换源图时重置裁切并保留 `portrait` 回退
- [x] 3.5 增加裁切校验和可定位错误，阻止非有限数值、非正宽高、越界矩形及不满足 `5:7` 像素比例容差的数据保存
- [x] 3.6 增加配置工具组件测试，覆盖上传后的默认框、拖动/缩放转换、重置、实时预览、保存读回、非法裁切定位和旧头像回退

## 4. Go 运行时快照

- [x] 4.1 在 matchengine 中定义可选 `ImageCrop` 值对象及统一合法性判断，并为 `PlayerProfile`、`PlayerState` 增加 `CardImage` 和 `AvatarCrop` JSON 字段
- [x] 4.2 扩展 `match.Service` 的 TbPlayer 适配器，在合法时将扁平 Luban 字段组装为 `AvatarCrop`，非法或缺失时保留 `Portrait` 且不输出可用裁切
- [x] 4.3 扩展回合状态投影与报告序列化，使卡面和裁切数据从 `PlayerProfile` 稳定复制到每个客户端 `PlayerState`
- [x] 4.4 增加 Go 单元测试，覆盖合法映射、缺失/非法回退、JSON 兼容、投影传递，以及仅改变视觉字段不会改变同种子模拟结果

## 5. 客户端选手视觉组件

- [x] 5.1 扩展比赛报告和页面所用 TypeScript 类型，表达可选 `cardImage`、`avatarCrop` 与现有 `portrait`
- [x] 5.2 实现共享 `PlayerFullCard` 组件，在稳定 `2:3` 容器中完整显示卡面并按 `cardImage -> portrait -> 默认图` 回退
- [x] 5.3 实现共享 `PlayerAvatarCrop` 组件，在固定 `5:7` 视口内依据原图自然尺寸和归一化矩形计算缩放/偏移，并按 `cardImage + crop -> portrait -> 默认图` 回退
- [x] 5.4 改造 TutorialPage 候选选手项，展示完整卡面、选手名、战队、价格和选中状态，同时保持现有预算、档位、人数与重复选择规则
- [x] 5.5 改造 BattlePage、TeamRoster 和 PlayerCard 的数据映射与头像渲染，使死亡灰化等状态样式叠加在动态裁切结果之上
- [x] 5.6 增加客户端测试，覆盖完整卡面选择、响应式 `2:3` 布局、精确 `5:7` 裁切、图片加载失败回退及回放状态变化时裁切保持不变

## 6. 使用说明与兼容迁移

- [x] 6.1 更新配置编辑器使用说明，记录完整卡面上传、裁切、重置、保存、写入 Luban 和旧 `portrait` 回退流程
- [x] 6.2 记录 Player 新字段的取值范围、`5:7` 像素比例公式、资源目录和迁移顺序，并明确系统只保存源图与参数、不生成裁切副本

## 7. 集成验证

- [x] 7.1 运行配置工具 `npm test`、`npm run lint` 和 `npm run build`，修复本变更引入的失败
- [x] 7.2 通过编辑器完成“上传完整卡面 -> 调整裁切 -> 保存 -> 写入 Luban -> 从项目中读取”的端到端验证，并确认 Excel 数值与读回画面一致
- [x] 7.3 再次运行 `scripts/gen-config.ps1`，确认 Luban 导表无错误且工作簿表头、已有数据和备份流程未被破坏
- [x] 7.4 运行服务端 `go test ./...`，确认新增快照字段、旧 Player 配置和确定性模拟测试全部通过
- [x] 7.5 运行客户端 `npm test`、`npm run lint` 和 `npm run build`，在桌面端与移动端核对 TutorialPage 完整卡面及 BattlePage 裁切头像无溢出遮挡
- [x] 7.6 按 README 使用 `nakama-pluginbuilder:3.30.0` Docker 镜像执行 `go build -buildmode=plugin -trimpath -o build/backend.so .`，并确认 Nakama 可重新加载生成的插件
