import { expect, test, type Browser, type Page } from "@playwright/test"

test.skip(!process.env.PLAYWRIGHT_LIVE_SOCIAL, "set PLAYWRIGHT_LIVE_SOCIAL=1 to run against the local Docker stack")
test.setTimeout(120_000)

async function bindFreshAccount(page: Page, email: string) {
  await page.goto("/")
  const onboarding = page.getByRole("button", { name: "先看看主页" })
  await onboarding.waitFor({ state: "visible", timeout: 8_000 }).catch(() => undefined)
  if (await onboarding.isVisible()) await onboarding.click()
  await page.getByRole("button", { name: "登录" }).click()
  await page.getByRole("button", { name: /第一次来/ }).click()
  const dialog = page.getByRole("dialog", { name: "绑定邮箱继续玩" })
  await dialog.getByRole("textbox", { name: "邮箱" }).fill(email)
  await dialog.locator('input[type="password"]').fill("Contact123!")
  await dialog.getByRole("button", { name: "绑定并保留进度" }).click()
  const code = await page.locator(".player-code span").textContent()
  expect(code).toMatch(/^玩家#[A-Z0-9]{8}$/)
  return code!.replace("玩家#", "")
}

async function openFriendsAndSaveProfile(page: Page, qq: string, wechat: string) {
  await page.getByRole("button", { name: "好友", exact: true }).click()
  await expect(page.getByRole("heading", { name: "好友与联系方式" })).toBeVisible()
  await expect(page.getByLabel("QQ")).toBeEnabled()
  await page.getByLabel("QQ").fill(qq)
  await page.getByLabel("微信号").fill(wechat)
  await page.getByRole("button", { name: "私密保存" }).click()
  await expect(page.getByText(/版本 1/)).toBeVisible()
}

async function refreshFriends(page: Page) {
  await page.getByRole("button", { name: "刷新" }).click()
}

async function runIsolatedPair(browser: Browser) {
  const contextA = await browser.newContext()
  const contextB = await browser.newContext()
  const pageA = await contextA.newPage()
  const pageB = await contextB.newPage()
  return { contextA, contextB, pageA, pageB }
}

test("two isolated accounts recover, accept, invalidate and reauthorize contact exchange", async ({ browser }) => {
  const { contextA, contextB, pageA, pageB } = await runIsolatedPair(browser)
  const unique = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  try {
    const codeA = await bindFreshAccount(pageA, `contact-a-${unique}@example.com`)
    const codeB = await bindFreshAccount(pageB, `contact-b-${unique}@example.com`)
    await openFriendsAndSaveProfile(pageA, "12345678", "contact_a")
    // B deliberately saves only QQ while A saves both channels. Acceptance must
    // authorize each side's own non-empty profile without requiring channel parity.
    await openFriendsAndSaveProfile(pageB, "87654321", "")

    await pageA.locator('.friend-add input').fill(`玩家#${codeB}`)
    await pageA.getByRole("button", { name: "添加好友" }).click()
    await refreshFriends(pageB)
    await expect(pageB.getByRole("heading", { name: "收到的好友请求" })).toBeVisible()
    await pageB.getByRole("button", { name: "接受好友请求" }).click()
    await refreshFriends(pageA)
    await expect(pageA.getByText(`玩家#${codeB}`).first()).toBeVisible()

    // B is offline when A requests. Returning to the page must recover the request from the inbox RPC.
    await pageB.goto("about:blank")
    await pageA.getByRole("button", { name: /联系方式$/ }).click()
    await pageA.getByRole("button", { name: /申请交换/ }).click()
    await expect(pageA.getByText(/即使消息卡片未送达/)).toBeVisible()
    await pageB.goto("http://localhost:3000/friends")
    await expect(pageB.getByRole("heading", { name: /联系方式申请 · 1 待处理/ })).toBeVisible()
    await pageB.getByRole("button", { name: "接受并互相授权" }).first().click()
    await expect(pageB.locator(".contact-nav-badge")).toHaveCount(0)

    await pageA.getByRole("button", { name: "关闭" }).click()
    await refreshFriends(pageA)
    await pageA.getByRole("button", { name: /联系方式$/ }).click()
    await expect(pageA.getByText("好友 QQ：")).toContainText("87654321")
    await expect(pageA.getByText("好友微信：")).toContainText("未提供")
    await pageA.getByRole("button", { name: "关闭" }).click()

    await pageB.getByRole("button", { name: /联系方式$/ }).click()
    await expect(pageB.getByText("好友 QQ：")).toContainText("12345678")
    await expect(pageB.getByText("好友微信：")).toContainText("contact_a")
    await pageB.getByRole("button", { name: "关闭" }).click()

    pageB.once("dialog", (dialog) => dialog.accept())
    await pageB.getByLabel("QQ").fill("87654322")
    await pageB.getByRole("button", { name: "私密保存" }).click()
    await expect(pageB.getByText(/版本 2/)).toBeVisible()

    await refreshFriends(pageA)
    await pageA.getByRole("button", { name: /联系方式$/ }).click()
    await expect(pageA.getByText("联系方式已变化，需要重新授权", { exact: true })).toBeVisible()
    await expect(pageA.getByText("87654321")).toHaveCount(0)
    await pageA.getByRole("button", { name: /重新申请交换/ }).click()
    const cooldown = pageA.getByText("操作过于频繁，请稍后再试。")
    const rateLimited = await cooldown.waitFor({ state: "visible", timeout: 2_000 }).then(() => true).catch(() => false)
    if (rateLimited) {
      await pageA.waitForTimeout(31_000)
      await pageA.getByRole("button", { name: /重新申请交换/ }).click()
    }
    await expect(pageA.getByText("交换请求等待处理")).toBeVisible()

    await refreshFriends(pageB)
    await pageB.getByRole("button", { name: "接受并互相授权" }).first().click()
    await pageA.getByRole("button", { name: "关闭" }).click()
    await refreshFriends(pageA)
    await pageA.getByRole("button", { name: /联系方式$/ }).click()
    await expect(pageA.getByText("好友 QQ：")).toContainText("87654322")
    await expect(pageA.getByText("好友微信：")).toContainText("未提供")

    const userBID = await pageB.evaluate(() => JSON.parse(atob(localStorage.getItem("nakama_token")!.split(".")[1])).uid as string)
    const security = await pageA.evaluate(async ({ friendID }) => {
      const token = localStorage.getItem("nakama_token")!
      const headers = { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }
      const read = await fetch("http://localhost:7350/v2/storage", {
        method: "POST", headers,
        body: JSON.stringify({ object_ids: [{ collection: "social_contact_profile", key: "profile", user_id: JSON.parse(atob(token.split(".")[1])).uid }] }),
      })
      const write = await fetch("http://localhost:7350/v2/storage", {
        method: "PUT", headers,
        body: JSON.stringify({ objects: [{ collection: "social_contact_profile", key: "profile", value: JSON.stringify({ qq: "forged-contact" }), permission_read: 2, permission_write: 1 }] }),
      })
      const forgedResponse = await fetch("http://localhost:7350/v2/rpc/SocialRespondContactExchange", {
        method: "POST", headers,
        body: JSON.stringify(JSON.stringify({ friend_id: friendID, request_id: "forged-request-id", accept: true })),
      })
      const inboxResponse = await fetch("http://localhost:7350/v2/rpc/SocialListContactExchangeInbox", {
        method: "POST", headers, body: JSON.stringify(JSON.stringify({ limit: 100 })),
      })
      return {
        readStatus: read.status,
        readBody: await read.text(),
        writeStatus: write.status,
        writeBody: await write.text(),
        forgedStatus: forgedResponse.status,
        forgedBody: await forgedResponse.text(),
        inboxBody: await inboxResponse.text(),
        localValues: Object.values(localStorage),
        toastText: [...document.querySelectorAll("[data-sonner-toast]")].map((node) => node.textContent ?? ""),
      }
    }, { friendID: userBID })
    expect(security.readStatus).toBe(200)
    expect(JSON.parse(security.readBody).objects ?? []).toHaveLength(0)
    expect(security.writeStatus).toBeGreaterThanOrEqual(400)
    expect(security.forgedStatus).toBeGreaterThanOrEqual(400)
    for (const serialized of [security.readBody, security.writeBody, security.forgedBody, security.inboxBody, ...security.localValues, ...security.toastText]) {
      expect(serialized).not.toContain("87654322")
    }

    await pageA.getByRole("button", { name: "关闭" }).click()
    const deleted = await pageA.evaluate(async (friendID) => {
      const token = localStorage.getItem("nakama_token")!
      return (await fetch(`http://localhost:7350/v2/friend?ids=${encodeURIComponent(friendID)}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } })).status
    }, userBID)
    expect([200, 204]).toContain(deleted)
    await refreshFriends(pageA)
    await refreshFriends(pageB)
    await expect(pageA.getByRole("heading", { name: "我的好友 · 0" })).toBeVisible()

    await pageA.locator('.friend-add input').fill(`玩家#${codeB}`)
    await pageA.getByRole("button", { name: "添加好友" }).click()
    await refreshFriends(pageB)
    await pageB.getByRole("button", { name: "接受好友请求" }).click()
    await refreshFriends(pageA)
    await pageA.getByRole("button", { name: /联系方式$/ }).click()
    await expect(pageA.getByText("授权已撤销", { exact: true })).toBeVisible()
    await expect(pageA.getByText("87654322")).toHaveCount(0)

    expect(codeA).not.toBe(codeB)
  } finally {
    await Promise.allSettled([contextA.close(), contextB.close()])
  }
})
