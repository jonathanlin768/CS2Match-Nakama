import assert from "node:assert/strict"
import test from "node:test"
import { ensurePagesProductionBranch } from "../scripts/configure-pages-project.mjs"

function projectResponse(productionBranch, status = 200) {
  return new Response(JSON.stringify({
    success: status >= 200 && status < 300,
    result: status >= 200 && status < 300 ? { production_branch: productionBranch } : null,
  }), {
    status,
    headers: { "content-type": "application/json" },
  })
}

test("Pages production branch already set to master is only verified", async () => {
  const calls = []
  const result = await ensurePagesProductionBranch({
    accountId: "account-id",
    apiToken: "api-token",
    projectName: "cs2match",
    fetchImpl: async (url, options = {}) => {
      calls.push({ url, options })
      return projectResponse("master")
    },
  })

  assert.deepEqual(result, { previousBranch: "master", productionBranch: "master" })
  assert.equal(calls.length, 2)
  assert.ok(calls.every(({ options }) => options.method === undefined))
})

test("Pages production branch is updated and verified before deployment", async () => {
  const calls = []
  let productionBranch = "main"
  const result = await ensurePagesProductionBranch({
    accountId: "account-id",
    apiToken: "api-token",
    projectName: "cs2match",
    targetBranch: "master",
    fetchImpl: async (url, options = {}) => {
      calls.push({ url, options })
      if (options.method === "PATCH") {
        productionBranch = JSON.parse(options.body).production_branch
      }
      return projectResponse(productionBranch)
    },
  })

  assert.deepEqual(result, { previousBranch: "main", productionBranch: "master" })
  assert.equal(calls.length, 3)
  assert.equal(calls[1].options.method, "PATCH")
  assert.deepEqual(JSON.parse(calls[1].options.body), { production_branch: "master" })
})

test("Pages API failures do not expose its token", async () => {
  const apiToken = "never-print-this-token"
  await assert.rejects(
    ensurePagesProductionBranch({
      accountId: "account-id",
      apiToken,
      projectName: "cs2match",
      fetchImpl: async () => { throw new Error(`network failure with ${apiToken}`) },
    }),
    (error) => {
      assert.match(error.message, /project lookup request failed/)
      assert.doesNotMatch(error.message, new RegExp(apiToken))
      return true
    },
  )
})
