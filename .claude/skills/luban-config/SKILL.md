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

