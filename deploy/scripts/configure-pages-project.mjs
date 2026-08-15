import { pathToFileURL } from "node:url"

function requireValue(value, name) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${name} is required`)
  }
  return value.trim()
}

async function requestProject(fetchImpl, url, options, operation) {
  let response
  try {
    response = await fetchImpl(url, options)
  } catch {
    throw new Error(`Cloudflare Pages ${operation} request failed`)
  }

  let payload
  try {
    payload = await response.json()
  } catch {
    throw new Error(`Cloudflare Pages ${operation} returned invalid JSON`)
  }

  if (!response.ok || payload?.success !== true || !payload.result) {
    throw new Error(`Cloudflare Pages ${operation} failed with HTTP ${response.status}`)
  }
  return payload.result
}

export async function ensurePagesProductionBranch({
  accountId,
  apiToken,
  projectName,
  targetBranch = "master",
  fetchImpl = fetch,
}) {
  const account = requireValue(accountId, "CLOUDFLARE_ACCOUNT_ID")
  const token = requireValue(apiToken, "CLOUDFLARE_API_TOKEN")
  const project = requireValue(projectName, "CLOUDFLARE_PAGES_PROJECT")
  const target = requireValue(targetBranch, "target production branch")
  const url = `https://api.cloudflare.com/client/v4/accounts/${encodeURIComponent(account)}/pages/projects/${encodeURIComponent(project)}`
  const headers = {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  }

  const current = await requestProject(fetchImpl, url, { headers }, "project lookup")
  const previousBranch = current.production_branch
  if (previousBranch !== target) {
    await requestProject(fetchImpl, url, {
      method: "PATCH",
      headers,
      body: JSON.stringify({ production_branch: target }),
    }, "production branch update")
  }

  const verified = await requestProject(fetchImpl, url, { headers }, "project verification")
  if (verified.production_branch !== target) {
    throw new Error(`Cloudflare Pages production branch is not ${target}`)
  }
  return { previousBranch, productionBranch: verified.production_branch }
}

async function main() {
  const targetBranch = process.argv[2] ?? "master"
  const result = await ensurePagesProductionBranch({
    accountId: process.env.CLOUDFLARE_ACCOUNT_ID,
    apiToken: process.env.CLOUDFLARE_API_TOKEN,
    projectName: process.env.CLOUDFLARE_PAGES_PROJECT,
    targetBranch,
  })
  if (result.previousBranch === result.productionBranch) {
    console.log(`[cs2match] Pages production branch verified: ${result.productionBranch}`)
  } else {
    console.log(`[cs2match] Pages production branch updated from ${result.previousBranch || "unset"} to ${result.productionBranch}`)
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch((error) => {
    console.error(error instanceof Error ? error.message : String(error))
    process.exitCode = 1
  })
}
