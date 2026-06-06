# Luban Excel 表格式完整文档（v4.5.0 兼容版）

> 适用于 **focus-creative-games/luban Next/4.x 系列**（确认兼容 v4.5.0+998c69c）。citeweb_search:14#10
> 基础格式参考官方文档与示例工程，4.5.0 特有特性已标注。

---

## 一、目录结构

```
Datas/
├── __tables__.xlsx      ← 表定义（table schema）
├── __beans__.xlsx       ← 结构定义（bean schema）
├── __enums__.xlsx       ← 枚举定义（enum schema）
├── #item.xlsx           ← 业务数据表（以 # 开头表示数据表）
├── #role.xlsx
└── reward.xlsx          ← 也可以不以 # 开头，只要在 __tables__.xlsx 里注册
```

---

## 二、元数据表格式（Schema 定义）

### 2.1 __tables__.xlsx — 表定义

定义你的业务数据表应该如何导出。

| 字段 | 说明 | 示例 |
|------|------|------|
| **name** | 表名（生成代码中的类名） | `TbItem` |
| **value** | 值类型（通常是 bean 名） | `Item` |
| **index** | 主键字段 | `id` |
| **mode** | 表模式 | `map` / `list` / `singleton` |
| **input** | 输入 Excel 文件名 | `item.xlsx` |
| **groups** | 所属分组 | `c,s` |
| **comment** | 注释 | 道具表 |
| **defineFromFile** | 是否从文件推断结构 | `true` / `false` |

**示例：**

| name | value | index | mode | input | groups | comment |
|------|-------|-------|------|-------|--------|---------|
| TbItem | Item | id | map | #item.xlsx | c,s | 道具表 |
| TbReward | Reward | | list | reward.xlsx | s | 奖励列表 |

> **注意：** `input` 的文件名如果带 `#` 前缀（如 `#item.xlsx`），表示该文件在 `dataDir` 下；不带 `#` 则按普通路径解析。citeweb_search:10#1

> **v4.5.0 新增：** 自动导入表时，文件名支持 `-<注释>` 后缀，该注释会作为 table 的 comment。例如 `item-道具表.xlsx` 会自动将 "道具表" 作为注释。citeweb_search:14#10

---

### 2.2 __beans__.xlsx — 结构定义

定义复合数据结构（Bean）。

| 字段 | 说明 | 示例 |
|------|------|------|
| **name** | 结构名 | `Item` |
| **fields** | 字段列表（逗号分隔） | `item_id,num,desc` |
| **comment** | 注释 | 道具结构 |
| **alias** | 别名 | 道具 |
| **sep** | 字段分隔符 | `,` / `\|` |

**更标准的写法（逐行定义字段）：**

| name | field | type | comment |
|------|-------|------|---------|
| Item | | | 道具结构 |
| | item_id | int | 道具ID |
| | num | int | 数量 |
| | desc | string | 描述 |

或者使用多行字段定义（每行一个字段）：

| name | field | type | comment |
|------|-------|------|---------|
| Reward | | | 奖励结构 |
| | item_id | int | 道具ID |
| | num | int | 数量 |

---

### 2.3 __enums__.xlsx — 枚举定义

定义枚举类型。

| 字段 | 说明 | 示例 |
|------|------|------|
| **name** | 枚举名 | `Color` |
| **unique** | 值是否唯一 | `true` |
| **flags** | 是否为标志位枚举 | `false` |
| **comment** | 注释 | 颜色枚举 |
| **items** | 枚举项列表 | 见下 |

**枚举项定义格式：**

| name | value | alias | comment |
|------|-------|-------|---------|
| Color | | | 颜色枚举 |
| RED | 1 | 红 | 红色 |
| GREEN | 2 | 绿 | 绿色 |
| BLUE | 3 | 蓝 | 蓝色 |

> **flags=true** 时，枚举项可以组合使用，如 `READ|EXECUTE`。citeweb_search:10#4
> 自 v3.1.3.0 起，可自定义分隔符：`Enums.emLifeTime#sep=,` 表示用逗号代替 `|`。citeweb_search:10#9

> **v4.5.0 新增：** 支持**常量别名**。可以在常量定义中使用别名，增强可读性。citeweb_search:14#10

---

## 三、数据表格式（业务 Excel）

### 3.1 表头三行规则

所有数据表必须以 **三行表头** 开头：

| ##var | id | name | price | rewards | ... |
|-------|----|------|-------|---------|-----|
| ##type | int | string | int | list,Reward | ... |
| ## | 道具ID | 名称 | 价格 | 奖励列表 | ... |
| | 1001 | 铁剑 | 500 | | ... |
| | 1002 | 木盾 | 300 | | ... |

**说明：**
- **第 1 行 (`##var`)**：字段名（对应 bean 或 table 的字段名）
- **第 2 行 (`##type`)**：字段类型（见下方类型系统）
- **第 3 行 (`##`)**：注释/描述（可选，用于策划阅读）
- **第 4 行起**：实际数据

> 第 1 列固定为 `##var`，表示这是 Luban 数据表。如果 A1 是 `##column`，则表示**纵表模式**（单例表常用）。citeweb_search:10#6

> **v4.5.0 修复：** 非数据单元簿（A1 单元不以 `##` 开头）的 A1 单元格不是字符串类型时，不再抛出异常。citeweb_search:14#10

---

### 3.2 类型系统

| 类型 | Excel 填写示例 | 说明 |
|------|---------------|------|
| `int` | `100` | 整数 |
| `long` | `10000000000` | 长整数 |
| `float` | `3.14` | 单精度浮点 |
| `double` | `3.1415926` | 双精度浮点 |
| `bool` | `true` / `false` | 布尔 |
| `string` | `hello` | 字符串 |
| `datetime` | `2024-01-01 12:00:00` | 日期时间 |
| `BeanName` | 见结构填写 | 自定义结构 |
| `EnumName` | `RED` / `红` / `1` | 枚举（可填名、别名或值） |
| `list,T` | 见列表填写 | 列表 |
| `set,T` | 同 list | 集合（去重） |
| `map,K,V` | 见 Map 填写 | 字典 |
| `T?` | 留空或 `null` | 可空类型 |

> **v4.5.0 新增：** 解析 Excel 数据时支持 **`0x` 前缀的16进制整数**。例如 `0xFF`（255）、`0x1A3F`（6719）都会被正确解析为 int/long 类型。citeweb_search:14#10

---

### 3.3 结构（Bean）填写

**方式一：限定列格式（推荐）**

通过合并单元格限定字段占用的列范围：

| ##var | id | name | item | | | | | |
|-------|----|------|------|-|-|-|-|-|
| ##type | int | string | Item | | | | | |
| ##var | | | item_id | num | desc | | | |
| | 1 | 铁剑 | 1001 | 1 | 新手剑 | | | |

> `Item` 结构占 3 列（item_id, num, desc），子字段用 `##var` 行声明。

**方式二：流式格式（紧凑）**

直接在单元格内按顺序填写：

| ##var | id | name | item |
|-------|----|------|------|
| ##type | int | string | Item |
| | 1 | 铁剑 | 1001,1,新手剑 |

> 默认用逗号分隔，可通过 `Item#sep=\|` 修改分隔符。citeweb_search:10#12

> **v4.5.0 新增：** type 的 attrs 支持**转义字符**，允许定义 sep 的分割符为 `&` 和 `#` 等特殊字符。例如 `Item#sep=&#` 表示用 `&` 和 `#` 组合分隔。citeweb_search:14#10

---

### 3.4 列表（list）填写

**限定列格式（横向展开）：**

| ##var | id | rewards | | | | | |
|-------|----|---------|-|-|-|-|-|
| ##type | int | list,Reward | | | | | |
| ##var | | 0 | 1 | 2 | | | |
| ##var | | item_id | num | desc | item_id | num | desc |
| | 1 | 1001 | 10 | 金币 | 1002 | 5 | 钻石 |

**流式格式（紧凑）：**

| ##var | id | rewards |
|-------|----|---------|
| ##type | int | list,Reward |
| | 1 | 1001,10,金币;1002,5,钻石 |

> 列表元素用 `;` 或换行分隔，元素内部用 `,` 或自定义 sep 分隔。citeweb_search:10#6

> **v4.5.0 修复：** 修复 `[*items,,,*items]` 格式的多列限定标题头没有处理 `*xxx` 多行数据标记的 bug。现在多行数据标记在多列限定格式中正常工作。citeweb_search:14#10

---

### 3.5 Map 填写

| ##var | id | extra |
|-------|----|-------|
| ##type | int | map,int,string |
| | 1 | 1,血量;2,攻击;3,防御 |

> 键值对依次填写，用 `;` 分隔每对，内部用 `,` 分隔 key 和 value。

---

### 3.6 可空类型（Nullable）

| ##var | id | item |
|-------|----|------|
| ##type | int | Item? |
| | 1 | {}1001,1,desc |
| | 2 | null |
| | 3 | |

> - 非空 bean 以 `{}` 开头，后面接字段值
> - `null` 或留空表示 null
> - 原子类型（int?）留空也表示 nullciteweb_search:10#15

---

### 3.7 多态 Bean（Polymorphic）

假设 `Shape` 有子类 `Circle`（radius）和 `Rectangle`（width, height）：

| ##var | id | shape | | |
|-------|----|-------|-|-|
| ##type | int | Shape | | |
| ##var | | $type | radius | width |
| | 1 | Circle | 10 | |
| | 2 | Rectangle | | 100 |
| | 3 | 圆 | 20 | |

> - `$type` 列指定具体子类名（或 alias）
> - 子类字段按顺序填写，不用的列留空
> - 如果 `$type` 填 `null` 或留空，表示 null（如果类型是 `Shape?`）citeweb_search:10#10

---

### 3.8 单例表（Singleton）

全局只有一份的配置，如系统参数。

**横表模式：**

| ##var | guild_open_level | bag_init_capacity | bag_max_capacity |
|-------|------------------|-------------------|------------------|
| ##type | int | int | int |
| ## | 公会开启等级 | 背包初始容量 | 背包最大容量 |
| | 10 | 100 | 500 |

**纵表模式（A1 = `##column`）：**citeweb_search:10#6

| ##var#column | ##type | ## | |
|--------------|--------|----|-|
| guild_open_level | int | 公会开启等级 | 10 |
| bag_init_capacity | int | 背包初始容量 | 100 |
| bag_max_capacity | int | 背包最大容量 | 500 |

---

### 3.9 多主键表

**联合索引（多个 key 组合唯一）：**

```xml
<table name="TbUnionMultiKey" value="UnionMultiKey" index="key1+key2+key3" input="union_multi_key.xlsx"/>
```

| ##var | key1 | key2 | key3 | num |
|-------|------|------|------|-----|
| ##type | int | long | string | int |
| | 1 | 2 | aaa | 123 |

**独立索引（多个 key 各自唯一）：**

```xml
<table name="TbMultiKey" value="MultiKey" index="key1,key2,key3" input="multi_key.xlsx"/>
```

| ##var | key1 | key2 | key3 | num |
|-------|------|------|------|-----|
| ##type | int | long | string | int |
| | 1 | 2 | aaa | 123 |

> 联合索引用 `+` 连接，独立索引用 `,` 连接。citeweb_search:10#15

---

### 3.10 无主键表（纯列表）

```xml
<table name="TbNotKeyList" value="NotKeyList" mode="list" input="not_key_list.xlsx"/>
```

| ##var | name | desc | value |
|-------|------|------|-------|
| ##type | string | string | int |
| | 记录1 | 描述 | 100 |
| | 记录2 | 描述 | 200 |

> `mode="list"` 且 `index` 为空，表示无主键。citeweb_search:10#15

---

## 四、紧凑格式（Compact Format）

自 v4.0.0 起，支持多种紧凑格式，在字段标题头用 `format` 属性指定：

| 格式 | 说明 | 示例 |
|------|------|------|
| **stream** | 流式格式（默认） | `1001,1,desc;2001,2,desc2` |
| **lite** | 简洁格式（带边界） | `(1001,1,desc),(2001,2,desc2)` |
| **json** | JSON 格式 | `[{"item_id":1001,"num":1}]` |
| **lua** | Lua 格式 | `{item_id=1001,num=1}` |

**指定方式：**

| ##var | id | rewards |
|-------|----|---------|
| ##type | int | list,Reward |
| | 1 | (1001,1,desc1),(2002,2,desc2) |

> 在字段名上定义：`rewards#format=lite`citeweb_search:10#12

---

## 五、字段属性（Field Attributes）

在字段名或类型后添加 `#key=value` 属性：

| 属性 | 作用 | 示例 |
|------|------|------|
| `default=xxx` | 默认值 | `x2#default=-1` |
| `sep=,` | 自定义分隔符 | `list,int#sep=,` |
| `format=lite` | 紧凑格式 | `rewards#format=lite` |
| `ref=TableName.field` | 引用检查 | `item_id#ref=TbItem.id` |
| `range=min~max` | 范围检查 | `price#range=0~10000` |
| `path` | 资源路径检查 | `icon#path` |

**示例：**

| ##var | id | price#default=0 | item_id#ref=TbItem.id |
|-------|----|-----------------|----------------------|
| ##type | int | int | int |
| | 1 | | 1001 |
| | 2 | 500 | 1002 |

> `price` 留空时默认值为 0，`item_id` 必须在 `TbItem` 表的 `id` 字段中存在。citeweb_search:10#10

> **v4.5.0 新增：** `path` 路径校验支持 **godot** 引擎。citeweb_search:14#10

---

## 六、Flags 枚举的特殊列限定

对于 `flags=1` 的枚举，可以用枚举项作为列名，非 0/非空列做或运算：

| ##var | type | | | | |
|-------|------|-|-|-|-|
| ##var | A | B | C | D | |
| | | 1 | 1 | | |
| | 1 | | | 1 | |

> 第一行数据：`A=0, B=1, C=1, D=0` → 值为 `B|C`
> 第二行数据：`A=1, B=0, C=0, D=1` → 值为 `A|D`citeweb_search:10#10

---

## 七、快速参考：最小可用示例

### 7.1 __enums__.xlsx

| name | unique | flags | comment | | | |
|------|--------|-------|---------|-|-|-|
| Color | true | false | 颜色 | | | |
| | name | value | alias | comment | | |
| | RED | 1 | 红 | 红色 | | |
| | GREEN | 2 | 绿 | 绿色 | | |
| | BLUE | 3 | 蓝 | 蓝色 | | |

### 7.2 __beans__.xlsx

| name | field | type | comment |
|------|-------|------|---------|
| Reward | | | 奖励结构 |
| | item_id | int | 道具ID |
| | num | int | 数量 |
| | desc | string | 描述 |

### 7.3 __tables__.xlsx

| name | value | index | mode | input | groups | comment |
|------|-------|-------|------|-------|--------|---------|
| TbItem | Item | id | map | #item.xlsx | c,s | 道具表 |

### 7.4 item.xlsx（数据表）

| ##var | id | name | price | reward | | |
|-------|----|------|-------|--------|-|-|
| ##type | int | string | int | Reward | | |
| ##var | | | | item_id | num | desc |
| | 1001 | 铁剑 | 500 | 2001 | 1 | 新手礼包 |
| | 1002 | 木盾 | 300 | 2002 | 2 | 防御礼包 |

---

## 八、v4.5.0 新特性速查表

| 特性 | 说明 | 影响范围 |
|------|------|---------|
| **0x 前缀16进制整数** | Excel 中可直接写 `0xFF` 表示 255 | int/long 字段填写 |
| **常量别名** | 常量定义支持别名，增强可读性 | `__enums__.xlsx` / `__beans__.xlsx` |
| **转义字符支持** | sep 分隔符可用 `&` 和 `#` 等特殊字符 | 流式格式字段 |
| **自动导入表注释** | 文件名 `-注释` 后缀自动作为 table comment | `__tables__.xlsx` 自动导入 |
| **多列限定修复** | `[*items,,,*items]` 格式支持 `*xxx` 多行标记 | 多列限定标题头 |
| **非数据簿容错** | A1 不以 `##` 开头且非字符串时不再异常 | 所有 Excel 文件 |
| **godot 路径校验** | `path` 属性支持 godot 资源路径 | 字段属性 |

---

## 九、相关链接

- 官方文档：https://www.datable.cn/docs/manual/excel
- 官方示例：https://github.com/focus-creative-games/luban_examples
- Excel 紧凑格式：https://www.datable.cn/docs/manual/excelcompactformat
- Release 页面：https://github.com/focus-creative-games/luban/releases
