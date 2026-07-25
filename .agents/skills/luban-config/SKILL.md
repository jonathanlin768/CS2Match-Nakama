---
name: luban-config
description: Guide for creating and exporting Luban config tables in the CS2Match project. Use when the user wants to add, modify, or export game configuration tables.
license: MIT
compatibility: Requires Docker Desktop and PowerShell on Windows.
metadata:
  author: CS2Match
  version: "1.0"
---

# Luban 配表技能（CS2Match 项目）

Use this skill when the user asks to add, modify, or export Luban configuration tables for CS2Match.

---

## When to use

- Adding a new config table (e.g., players, routes, weapons, combat constants).
- Modifying existing config table schema or data.
- Exporting config tables to Go / TypeScript code.
- Troubleshooting Luban export errors.

---

## Project structure

| Path | Purpose |
|---|---|
| `configs/luban.conf` | Luban main config with groups `c`/`s`/`e` and targets `client`/`server`/`all`. |
| `configs/Datas/` | Root data directory for all Excel tables. |
| `configs/Datas/__tables__.xlsx` | Manual table registration. **Only use for files without `#` prefix.** |
| `configs/Datas/__beans__.xlsx` | Optional bean definitions. |
| `configs/Datas/__enums__.xlsx` | Enum definitions. |
| `scripts/gen-config.ps1` | Export script (Windows + Docker). |
| `server/config/` | Generated Go code + embedded JSON data. |
| `client/src/config/` | Generated TypeScript code. |
| `client/public/data/config/` | Generated client JSON data. |

---

## Rules

### 1. Prefer auto-import naming

Create Excel files as `configs/Datas/#<Name>.xlsx`.

| File | Generated table | Generated bean | JSON data |
|---|---|---|---|
| `#Player.xlsx` | `TbPlayer` | `Player` | `tbplayer.json` |
| `#Item.xlsx` | `TbItem` | `Item` | `tbitem.json` |
| `#Route.xlsx` | `TbRoute` | `Route` | `tbroute.json` |

- Do **NOT** register `#<Name>.xlsx` files in `__tables__.xlsx`. Auto-import already handles them.
- Manual registration is only for files **without** the `#` prefix.

### 2. Use the standard 4-row header

Every data table must start with these exact rows. Data rows begin at row 5. Column A must be empty for data rows.

```text
| ##var  | id     | name   | entry | aim  | ... |
| ##type | string | string | int   | int  | ... |
| ##group|        | c,s,e  | c,s,e | c,s,e| ... |
| ##     | 选手ID | 名称   | 突破  | 精准 | ... |
|        | player_001 | NiKo | 85  | 92   | ... |
|        | player_002 | s1mple | 90 | 88   | ... |
```

- `##var`: field names.
- `##type`: field types.
- `##group`: export groups per field (`c`, `s`, `e`; comma separated). Leave `id` and common fields empty.
- `##`: Chinese comments for designers.
- Data rows: **A column must be empty**; values start at column B.

### 3. Common field types

| Type | Example | Notes |
|---|---|---|
| `int` | `85` | Integer. |
| `string` | `NiKo` | String. |
| `bool` | `true` | Boolean. |
| `float` | `1.5` | Float. |
| `list,string` | `Entry,AWPer` | List of strings. |
| `(list#sep=,),string` | `Entry,AWPer` | Explicit comma-separated string list. |
| `list,int` | `1,2,3` | List of integers. |
| `BeanName` | `Reward` | Reference a bean from `__beans__.xlsx` or another table. |
| `EnumName` | `A` / `红` / `1` | Enum value can be name, alias, or numeric value. |

### 3.1 Complex types from local Luban examples

The local Luban examples in `D:\Project\luban_examples\DataTables\Data` show how to define and fill beans, enums, maps, and nested collections. Use these as the reference when a CS2Match table needs more than primitive fields.

#### Bean definitions in `__beans__.xlsx`

`configs/Datas/__beans__.xlsx` uses a two-level header. A bean starts on the row where `full_name` is filled. Its fields are listed from the `*fields` columns on that row and following rows until the next `full_name`.

```text
| ##var | full_name | parent | valueType | sep | alias | comment | group | tags | *fields |       | type   | group | comment |
| ##var |           |        |           |     |       |         |       |      | name    | alias | type   | group | comment |
| ##    | ...       | ...    | ...       | ... | ...   | ...     | ...   | ...  | field   | alias | type   | group | comment |
|       | Reward    |        | 1         | ,   |       | reward  | c,s,e |      | item_id |       | string | c,s,e | item id |
|       |           |        |           |     |       |         |       |      | count   |       | int    | c,s,e | count   |
```

- `full_name`: bean type name. Names may include a module prefix in examples, such as `test.DemoE2`.
- `parent`: optional parent bean for inheritance or polymorphic beans.
- `valueType`: optional. Examples use `1` for value-like beans.
- `sep`: optional default separator for parsing compact bean cell values, for example `vec3` uses `sep=","`.
- Field columns under `*fields`: `name`, `alias`, `type`, `group`, `comment`, `tags`, `variants`.

Example from `Defines/builtin.xml`:

```xml
<bean name="vec3" valueType="1" sep=",">
    <var name="x" type="float"/>
    <var name="y" type="float"/>
    <var name="z" type="float"/>
</bean>
```

With `sep=","`, a table cell for `vec3` can be written as `1,2,3`.

#### Enum definitions in `__enums__.xlsx`

`configs/Datas/__enums__.xlsx` also uses a two-level header. An enum starts where `full_name` is filled. Enum items are listed from the `*items` columns on that row and following rows until the next `full_name`.

```text
| ##var | full_name | flags | unique | group | comment | tags | *items |       | value | comment |
| ##var |           |       |        |       |         |      | name   | alias | value | comment |
| ##    | ...       | ...   | ...    | ...   | ...     | ...  | name   | alias | value | comment |
|       | Rarity    | false | true   | c,s,e | rarity  |      | C      | normal| 1     | common  |
|       |           |       |        |       |         |      | B      | rare  | 2     | rare    |
```

- `flags=true` creates a bit-flag enum. Example values can be combined with `|`, such as `WRITE|READ`.
- `unique=true` requires enum values to be unique.
- Data cells can use enum `name`, `alias`, or numeric `value`.

#### Map fields

Map type syntax is:

```text
map,<keyType>,<valueType>
```

Examples:

```text
map,int,int
map,long,int
map,string,int
map,EnumName,int
map,string,BeanName
```

When using compact cell values, the value is a flat sequence of key/value pairs. With comma separation:

```text
| ##var  | x1#sep=,    | x3#sep=,       |
| ##type | map,int,int | map,string,int |
|        | 1,2,3,4     | aaa,1,bbb,2    |
```

This exports as key/value entries equivalent to:

```json
[
  {"key": 1, "value": 2},
  {"key": 3, "value": 4}
]
```

Use parentheses when the container itself has parameters:

```text
(map#sep=,),int,int
(map#sep=,),int#ref=TbItem,int#ref=TbItem
(map#(size=(1, 3))),int,int
```

#### Nested collections

The local examples show nested containers and custom separators:

```text
(list#sep=|),(list#sep=,),int
(list#sep=|),(set#sep=,),int
(map#sep=-|),int,(list#sep=,),int
```

Example values:

```text
| ##type | (list#sep=|),(list#sep=,),int | (map#sep=-|),int,(list#sep=,),int |
|        | 1,2|3,4,5,6                   | 1-1,2,3|2-2,4,6                   |
```

Rules of thumb:

- Wrap a container in parentheses when adding parameters (`#sep`, `#size`, `#index`) or when nesting it inside another container.
- Prefer explicit separators for nested data. Reusing `,` at every level becomes hard to read and easy to break.
- Simple wrappers such as `list,(int#ref=TbX)` are used in examples for element-level parameters.
- Multi-dimensional collection examples exist in `test/multi_dimension.xlsx`, but the related table is commented out in the example XML. For CS2Match, run a small export validation before relying on a new nested shape in production tables.
### 4. Export command

Ensure Docker Desktop is running, then run from project root:

```powershell
powershell -File scripts/gen-config.ps1
```

This builds the Luban Docker image, cleans old generated files, and regenerates Go + TypeScript code and JSON data.

### 5. Always update hand-written loader files after export

`gen-config.ps1` deletes and recreates these files if they are missing:

- `server/config/loader.go`
- `client/src/config/index.ts`

After adding a new table, update them:

**`server/config/loader.go`**: `Init()` loads all JSON files automatically. Add the new table to `TableCount()` and any helper functions.

```go
if Global.TbPlayer != nil {
    count++
}
```

**`client/src/config/index.ts`**: add the new table's JSON name (lowercase) to `TABLE_NAMES` and export the generated types.

```ts
export { Tables, Tbitem, item, TbPlayer, Player } from './schema';

const TABLE_NAMES = ['tbitem', 'tbplayer'];
```

### 6. Verify after export

Run both checks before reporting completion:

```bash
cd server/config
go test -v ./...
```

```bash
cd client
npx tsc --noEmit
```

---

## Manual registration (only for non-`#` files)

If a file is not auto-imported, add a row to `configs/Datas/__tables__.xlsx`:

```text
| full_name | value_type | read_schema_from_file | input            | index | mode | group | comment |
|-----------|------------|-----------------------|------------------|-------|------|-------|---------|
| TbPlayer  | Player     | true                  | tb_player.xlsx   | id    | map  | c,s,e | 选手表  |
```

- `input` without `#` is relative to `configs/Datas/`.
- `read_schema_from_file: true` means schema is inferred from the Excel header.

---

## Common pitfalls

| Symptom | Cause | Fix |
|---|---|---|
| `xxx is not int type` with field shifted | Data row has a value in column A. | Leave column A empty in data rows. |
| Same error even with empty A column | Missing `##group` row. | Add `##var` / `##type` / `##group` / `##` headers. |
| `type:'Player' duplicate` | Registered a `#Player.xlsx` file in `__tables__.xlsx`. | Remove manual registration for `#` files. |
| Client says table not found | New table JSON not loaded. | Add `tbxxx` to `client/src/config/index.ts` `TABLE_NAMES`. |
| `failed to connect to docker API` | Docker Desktop not running. | Start Docker Desktop and retry. |
| Hand-written loader logic lost | Luban deleted `loader.go`/`index.ts`; script recreated defaults. | Re-apply custom changes after export. |

---

## Example workflow

User: "Add a route config table."

1. Create `configs/Datas/#Route.xlsx`:

```text
| ##var  | id     | name    | target_site | base_time | min_players | max_players |
| ##type | string | string  | string      | int       | int         | int         |
| ##group|        | c,s,e   | c,s,e       | c,s,e     | c,s,e       | c,s,e       |
| ##     | 路线ID | 路线名  | 目标包点    | 基础时间  | 最少人数    | 最多人数    |
|        | A_Long | A大     | A           | 20        | 1           | 3           |
|        | A_Short| A小     | A           | 15        | 1           | 2           |
```

2. Run `powershell -File scripts/gen-config.ps1`.
3. Update `client/src/config/index.ts`:

```ts
export { Tables, Tbitem, item, TbRoute, Route } from './schema';
const TABLE_NAMES = ['tbitem', 'tbroute'];
```

4. Update `server/config/loader.go` `TableCount()` to include `TbRoute`.
5. Run `go test -v ./...` in `server/config` and `npx tsc --noEmit` in `client`.
6. Report results.

---

## References

- Official examples on this machine: `D:\Project\luban_examples-main`
- Project examples: `configs/Datas/#Player.xlsx`, `configs/Datas/#item.xlsx`
- Demo config archive for this project: `doc/luban-config-demo/Datas`

