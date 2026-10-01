## 1. 基线

- [x] 1.1 核对规约、原有改动，运行 Match / Social 基线测试。

## 2. 实现

- [x] 2.1 统一 engineService 命名，提取 MatchMode、preparedMatch 和模式准备函数。
- [x] 2.2 增加 RPC 注册助手和 Match 注册入口，整理 main 与 Social 注册清单；总注册函数逐个检查模块错误，全部完成后返回 nil。
- [x] 2.3 验证注册失败处理、模式非法值和 RPC 认证边界。
- [x] 2.4 更新项目结构与教程中的调用链、命名和注册说明。

## 3. 验证

- [x] 3.1 运行 Go 全量测试与受影响包 Linux race 检查。
- [x] 3.2 运行现有前端契约与回放检查。
- [x] 3.3 编译更新 Linux 插件，验证健康、认证、两个比赛模式、调试 RPC、权限错误和 Social hooks。
- [x] 3.4 严格校验 OpenSpec，检查差异并记录结果。
