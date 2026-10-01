## Context

Nakama 的 InitModule 是依赖组装入口。Match RPC 在 main 中逐个注册，Social 在模块内用 map 注册；两处重复注册逻辑。SimuMatch 已经共享模拟和响应构造，模式差异仅在配置校验与组队。当前三个 TODO 的调整已获用户确认，其他未提交的引擎重构与注释需要保留。

## Goals / Non-Goals

**Goals:**
- 用 `engineService` 区分长期服务依赖与单场运行时对象。
- 让模式选择只负责分派，将各模式准备步骤独立组织。
- 模块拥有注册清单，通用助手统一注册顺序、成功日志和错误包装。
- 保持网络协议、错误码、认证和教学阵容规则。

**Non-Goals:**
- 反射发现、注解、生成器、策略类层级、模式注册容器。
- 新玩法、RPC、Storage、Match Handler 或配表。

## Decisions

1. `MatchMode` 使用定义的 string 类型，常量值继续是 `computer` / `tutorial`。标准 JSON 解码保持未知字符串可解码，再由服务返回原有 INVALID_MODE；不新增自定义解码器。
2. `preparedMatch` 只包含 MapID 和两个 TeamInput。prepareMatch 的短 switch 分派到 prepareComputerMatch / prepareTutorialMatch；配置版本、预算和组队错误按原顺序返回。公共模拟流程保留原有时间、seed、错误转换与响应构造。
3. 使用 `internal/framework/rpcregistry` 共享有类型的 Entry / Handler 和 Register 函数。Registrar 接口只包含 Nakama 原始 RegisterRpc 签名，生产传入 runtime.Initializer。Register 按切片顺序注册，失败即停止，以 %w 包装 RPC 名称，成功后统一日志；权限继续由各 handler 检查。
4. Match 新增 RegisterRPCs 模块入口。main 保持依赖组装与 HealthCheck 注册；总注册函数逐个检查各模块的注册错误，全部完成后统一返回 nil，后续模块可继续按顺序添加。Social 保持现有 Register 入口与启用开关，将六个 RPC 改为切片并调用助手。关闭交换功能时仍注册原有好友删除和聊天拒绝 hooks。
5. 不引入模式 interface，两个私有准备函数已满足当前扩展和阅读需求。明确函数边界之后，可在模式拥有独立依赖时提取接口。

## Risks / Trade-offs

- [枚举被误解为封闭集合] → 保留服务端未知/空模式校验和错误码，并验证 JSON 仍使用字符串。
- [提取函数改变错误顺序或规则] → 保留校验与组队顺序，使用现有 computer / tutorial 集成测试覆盖配置、预算和双方身份。
- [注册整理丢失 handler、依赖或 hooks] → 验证有序注册失败处理与 Match handler 认证边界；加载新插件后调用实际健康、调试、正式人机和教学 RPC，并核对 Social hooks。
- [错误包装影响诊断] → 使用 %w 保留原错误，明确失败 RPC 名称，不在助手吞掉错误。

## Migration Plan

无数据或客户端迁移。运行 Go / 客户端现有测试，使用项目固定 pluginbuilder 构建，按本地更新 skill 重启 Nakama 并验证实际 HTTP RPC。构建失败时保留旧产物并停止更新。源码与构建产物同步回退即可恢复原布局。

## Open Questions

无。
