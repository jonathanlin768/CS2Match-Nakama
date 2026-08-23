import { expect, test, type Page, type Route } from "@playwright/test"

interface MockSocialState {
  profileFailures: number
  profile: { qq: string; wechat: string; revision: number; summary: { qq_configured: boolean; wechat_configured: boolean; qq_masked?: string; wechat_masked?: string } }
  inbox: {
    received: Array<Record<string, unknown>>
    sent: Array<Record<string, unknown>>
    reapproval: Array<Record<string, unknown>>
    incoming_pending_count: number
  }
  exchange: Record<string, unknown>
}

function jwt(payload: Record<string, unknown>) {
  const encode = (value: object) => Buffer.from(JSON.stringify(value)).toString("base64url")
  return `${encode({ alg: "HS256", typ: "JWT" })}.${encode(payload)}.mock-signature`
}

async function mockLoggedInSocial(page: Page, overrides: Partial<MockSocialState> = {}) {
  const now = Math.floor(Date.now() / 1000)
  // Auth restoration currently compares expiry against Date.now() milliseconds.
  const token = jwt({ uid: "user-a", usn: "AAAAAAAA", exp: Date.now() + 3_600_000 })
  const refreshToken = jwt({ uid: "user-a", usn: "AAAAAAAA", exp: Date.now() + 7_200_000 })
  await page.addInitScript(({ token: access, refresh }) => {
    localStorage.setItem("nakama_token", access)
    localStorage.setItem("nakama_refresh", refresh)
    localStorage.setItem("nakama_identity_kind", "account")
  }, { token, refresh: refreshToken })

  const state: MockSocialState = {
    profileFailures: 0,
    profile: {
      qq: "12345678",
      wechat: "wx_account",
      revision: 4,
      summary: { qq_configured: true, wechat_configured: true, qq_masked: "12***78", wechat_masked: "wx***nt" },
    },
    inbox: {
      received: [{ friend_id: "user-b", username: "BBBBBBBB", request_id: "request-1", requester_id: "user-b", recipient_id: "user-a", channels: ["qq", "wechat"], status: "pending", version: 1, requested_at: now - 60, expires_at: now + 3600 }],
      sent: [],
      reapproval: [],
      incoming_pending_count: 1,
    },
    exchange: { request_id: "request-1", requester_id: "user-b", recipient_id: "user-a", channels: ["qq", "wechat"], status: "pending", version: 1 },
    ...overrides,
  }

  const rpc = async (route: Route, id: string) => {
    if (id === "SocialGetContactProfile") {
      if (state.profileFailures > 0) {
        state.profileFailures -= 1
        await route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ code: 13, message: "INTERNAL: private=should-not-leak" }) })
        return
      }
      await route.fulfill({ json: { id, payload: JSON.stringify(state.profile) } })
      return
    }
    if (id === "SocialListContactExchangeInbox") {
      await route.fulfill({ json: { id, payload: JSON.stringify(state.inbox) } })
      return
    }
    if (id === "SocialGetContactExchange") {
      await route.fulfill({ json: { id, payload: JSON.stringify(state.exchange) } })
      return
    }
    if (id === "SocialRespondContactExchange") {
      const encodedBody = route.request().postDataJSON() as string
      const body = JSON.parse(encodedBody) as { accept: boolean }
      state.inbox = { received: [], sent: [], reapproval: [], incoming_pending_count: 0 }
      state.exchange = body.accept
        ? { request_id: "request-1", requester_id: "user-b", recipient_id: "user-a", channels: ["qq", "wechat"], status: "accepted", version: 2, friend_contact: { qq: "87654321", wechat: "friend_wx" } }
        : { request_id: "request-1", requester_id: "user-b", recipient_id: "user-a", channels: ["qq", "wechat"], status: "declined", version: 2 }
      await route.fulfill({ json: { id, payload: JSON.stringify(state.exchange) } })
      return
    }
    if (id === "SocialSetContactProfile") {
      const encodedBody = route.request().postDataJSON() as string
      const body = JSON.parse(encodedBody) as { qq: string; wechat: string }
      state.profile = {
        qq: body.qq.trim(),
        wechat: body.wechat.trim(),
        revision: state.profile.revision + 1,
        summary: { qq_configured: Boolean(body.qq.trim()), wechat_configured: Boolean(body.wechat.trim()), qq_masked: "98***10", wechat_masked: "wx***ew" },
      }
      state.exchange = { status: "stale", version: 3 }
      state.inbox = { received: [], sent: [], reapproval: [{ friend_id: "user-b", username: "BBBBBBBB", status: "stale", version: 3, channels: ["qq", "wechat"] }], incoming_pending_count: 0 }
      await route.fulfill({ json: { id, payload: JSON.stringify(state.profile) } })
      return
    }
    if (id === "SocialRequestContactExchange") {
      state.exchange = { request_id: "request-2", requester_id: "user-a", recipient_id: "user-b", channels: ["qq", "wechat"], status: "pending", version: 4 }
      state.inbox = { received: [], sent: [{ friend_id: "user-b", username: "BBBBBBBB", ...state.exchange }], reapproval: [], incoming_pending_count: 0 }
      await route.fulfill({ json: { id, payload: JSON.stringify(state.exchange) } })
      return
    }
    await route.fulfill({ status: 404, json: { message: "unknown RPC" } })
  }

  await page.route("http://localhost:7350/**", async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname === "/v2/friend" && route.request().method() === "GET") {
      await route.fulfill({ json: { friends: [{ state: 0, user: { id: "user-b", username: "BBBBBBBB", online: true } }] } })
      return
    }
    const match = /^\/v2\/rpc\/(.+)$/.exec(url.pathname)
    if (match) { await rpc(route, decodeURIComponent(match[1])); return }
    await route.abort()
  })
  return state
}

test("authoritative request group, badge and bidirectional acceptance work without DM history", async ({ page }) => {
  await mockLoggedInSocial(page)
  await page.goto("/friends")

  await expect(page.getByRole("heading", { name: /联系方式申请 · 1 待处理/ })).toBeVisible()
  await expect(page.getByLabel("1 条待处理联系方式申请")).toBeVisible()
  await expect(page.getByLabel("QQ")).toHaveValue("12345678")
  await expect(page.getByLabel("微信号")).toHaveValue("wx_account")
  const headings = page.locator(".friends-list h2")
  await expect(headings.first()).toContainText("联系方式申请")

  await page.getByRole("button", { name: "接受并互相授权" }).first().click()
  await expect(page.getByText("暂无联系方式申请")).toBeVisible()
  await expect(page.locator(".contact-nav-badge")).toHaveCount(0)

  await page.getByRole("button", { name: /联系方式$/ }).click()
  await expect(page.getByText("好友 QQ：")).toContainText("87654321")
  await expect(page.getByText("好友微信：")).toContainText("friend_wx")
})

test("profile retry preserves local edits and a real change requires invalidation confirmation", async ({ page }) => {
  const state = await mockLoggedInSocial(page, { profileFailures: 1 })
  await page.goto("/friends")

  await expect(page.getByText(/本地已编辑内容不会被清空/)).toBeVisible()
  await page.getByLabel("QQ").fill("9876543210")
  await page.getByLabel("微信号").fill("wx_local_edit")
  await page.getByRole("button", { name: /重试$/ }).click()
  await expect(page.getByLabel("QQ")).toHaveValue("9876543210")
  await expect(page.getByLabel("微信号")).toHaveValue("wx_local_edit")

  page.once("dialog", async (dialog) => {
    expect(dialog.message()).toContain("所有既有交换授权立即失效")
    await dialog.accept()
  })
  await page.getByRole("button", { name: "私密保存" }).click()
  await expect(page.getByText(/版本 5/)).toBeVisible()
  expect(state.profile.qq).toBe("9876543210")
})

test("received request can be declined directly from the authoritative group", async ({ page }) => {
  await mockLoggedInSocial(page)
  await page.goto("/friends")
  await page.getByRole("button", { name: "拒绝" }).first().click()
  await expect(page.getByText("暂无联系方式申请")).toBeVisible()
  await expect(page.locator(".contact-nav-badge")).toHaveCount(0)
})

for (const status of ["stale", "revoked"] as const) {
  test(`${status} state never renders contact payload and stays usable on mobile`, async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await mockLoggedInSocial(page, {
      inbox: { received: [], sent: [], reapproval: [{ friend_id: "user-b", username: "BBBBBBBB", status, version: 7, channels: ["qq", "wechat"] }], incoming_pending_count: 0 },
      exchange: { status, version: 7, friend_contact: { qq: "must-not-render", wechat: "must-not-render" } },
    })
    await page.goto("/friends")
    await page.getByRole("button", { name: /玩家#BBBBBBBB/ }).first().click()
    await expect(page.getByText(status === "stale" ? "联系方式已变化，需要重新授权" : "授权已撤销", { exact: true })).toBeVisible()
    await expect(page.getByText("must-not-render")).toHaveCount(0)
    await expect(page.getByRole("button", { name: /重新申请交换/ })).toBeVisible()
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth)
    expect(overflow).toBe(false)
  })
}
