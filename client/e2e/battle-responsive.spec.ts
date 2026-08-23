import { expect, test, type Page, type TestInfo } from "@playwright/test"
import { battleReportFixture } from "./fixtures/match-report"

async function openBattle(page: Page) {
  await page.goto("/")
  await page.evaluate((report) => {
    window.history.replaceState({ usr: { report }, key: "battle-e2e", idx: 0 }, "", "/battle")
    window.location.reload()
  }, battleReportFixture)
  await expect(page.locator("[data-battle-layout]")) .toBeVisible()
}

async function assertNoPageOverflow(page: Page) {
  const overflow = await page.evaluate(() => ({
    body: document.body.scrollWidth - document.body.clientWidth,
    root: document.documentElement.scrollWidth - document.documentElement.clientWidth,
  }))
  expect(overflow.body).toBeLessThanOrEqual(1)
  expect(overflow.root).toBeLessThanOrEqual(1)
}

async function capture(page: Page, testInfo: TestInfo, name: string) {
  await page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true })
}

test("390px portrait keeps the battle stage and auxiliary views usable", async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await openBattle(page)
  await expect(page.locator('[data-battle-layout="compact"]')).toBeVisible()
  await page.getByRole("button", { name: "暂停" }).click()
  await expect(page.locator('[data-testid="compact-live-stage"]')).toBeVisible()
  await expect(page.locator('[data-testid="alive-player"]')).toHaveCount(10)
  await assertNoPageOverflow(page)

  await page.getByRole("button", { name: "阵容", exact: true }).click()
  await expect(page.locator(".compact-player-row")).toHaveCount(5)
  await page.getByRole("button", { name: "数据统计", exact: true }).click()
  await expect(page.locator(".battle-stats-table")).toHaveCount(2)
  await page.getByRole("button", { name: "实时战况", exact: true }).click()
  await expect(page.locator('[data-testid="compact-live-stage"]')).toBeVisible()
  await page.getByRole("button", { name: "跳过比赛" }).click()
  await page.locator(".battle-round-button").filter({ hasText: /^1$/ }).click()
  await expect(page.locator('.battle-round-button[aria-current="step"]')).toHaveText("1")
  const scrollTop = await page.locator(".app-content").evaluate((element) => {
    element.scrollTop = element.scrollHeight
    return element.scrollTop
  })
  expect(scrollTop).toBeGreaterThan(0)
  await capture(page, testInfo, "battle-390x844")
})

test("layout switching preserves playback context", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await openBattle(page)
  await page.getByRole("button", { name: "暂停" }).click()
  await expect(page.locator('[data-battle-layout="compact"]')).toBeVisible()
  await page.setViewportSize({ width: 1440, height: 900 })
  await expect(page.locator('[data-battle-layout="desktop"]')).toBeVisible()
  await expect(page.getByText("第 1 回合", { exact: false }).first()).toBeVisible()
  await page.setViewportSize({ width: 844, height: 390 })
  await expect(page.locator('[data-battle-layout="compact"]')).toBeVisible()
  await expect(page.getByText("第 1 回合", { exact: false }).first()).toBeVisible()
  await assertNoPageOverflow(page)
})

for (const viewport of [
  { width: 390, height: 844, name: "mobile" },
  { width: 1024, height: 768, name: "desktop-boundary" },
  { width: 1440, height: 900, name: "desktop" },
]) {
  test(`radar markers stay in the shared SVG coordinate space at ${viewport.name}`, async ({ page }, testInfo) => {
    await page.setViewportSize(viewport)
    await openBattle(page)
    await page.getByRole("button", { name: "跳过比赛" }).click()
    const markers = page.locator('[data-testid="battle-map-marker"]')
    await expect(markers).toHaveCount(3)
    const alignment = await page.locator('[data-testid="battle-map"]').evaluate((container) => {
      const svg = container.querySelector("svg")!
      const marker = container.querySelector('[data-testid="battle-map-marker"][data-map-x="0.5"]') as SVGGElement
      const point = svg.createSVGPoint()
      point.x = 512
      point.y = 492
      const expected = point.matrixTransform(svg.getScreenCTM()!)
      const markerBox = marker.getBoundingClientRect()
      return { dx: Math.abs(markerBox.x + markerBox.width / 2 - expected.x), dy: Math.abs(markerBox.y + markerBox.height / 2 - expected.y) }
    })
    expect(alignment.dx).toBeLessThan(1)
    expect(alignment.dy).toBeLessThan(1)
    await assertNoPageOverflow(page)
    await capture(page, testInfo, `radar-${viewport.width}x${viewport.height}`)
  })
}

test("desktop layout retains rosters, bomb state, map and event feed", async ({ page }) => {
  await page.setViewportSize({ width: 1920, height: 900 })
  await openBattle(page)
  await expect(page.locator('[data-battle-layout="desktop"]')).toBeVisible()
  await expect(page.locator(".battle-roster")).toHaveCount(2)
  await expect(page.locator('[data-testid="bomb-status"]')).toBeVisible()
  await expect(page.locator('[data-testid="battle-map"]')).toBeVisible()
  await expect(page.locator('[data-testid="event-feed-list"]')).toBeVisible()
  await expect(page.locator('[data-battle-layout="compact"]')).toHaveCount(0)
  await assertNoPageOverflow(page)
})
