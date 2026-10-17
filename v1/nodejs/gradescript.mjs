// A small client for the GradeScript Grading API: https://console.gradescriptug.com/docs
// Node.js 20.12 or later, no dependencies (built-in fetch, FormData, openAsBlob and
// process.loadEnvFile).
import { openAsBlob } from 'node:fs'
import { basename } from 'node:path'

// Read GRADESCRIPT_API_KEY (and GRADESCRIPT_API_URL) from the .env file at the repo root.
// Variables already set in the environment take precedence.
try {
  process.loadEnvFile(new URL('../../.env', import.meta.url))
} catch (error) {
  if (error.code !== 'ENOENT') throw error
}

const API_URL = (
  process.env.GRADESCRIPT_API_URL || 'https://api.gradescriptug.com/client/v1'
).replace(/\/+$/, '')

// A grading is finished once it reaches one of these statuses.
export const FINISHED = ['COMPLETED', 'PARTIAL', 'FAILED']

// GETs are safe to repeat, so they are retried after these delays on network
// errors, 429 and 5xx. POST and PATCH are never retried: creating a grading
// twice would grade, and bill, the script twice.
const GET_RETRY_DELAYS_MS = [1000, 2000, 4000]

// An error response from the API: { status, error, message }.
export class GradeScriptError extends Error {
  constructor(statusCode, error, message) {
    super(`HTTP ${statusCode} ${error}: ${message}`)
    this.statusCode = statusCode
    this.error = error
  }
}

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

async function request(method, path, { query = {}, body } = {}) {
  const apiKey = process.env.GRADESCRIPT_API_KEY
  if (!apiKey) {
    throw new Error('Set GRADESCRIPT_API_KEY in the .env file at the repo root (see .env.example).')
  }

  const url = new URL(API_URL + path)
  for (const [name, value] of Object.entries(query)) {
    if (value !== undefined) url.searchParams.set(name, value)
  }

  const delays = method === 'GET' ? GET_RETRY_DELAYS_MS : []
  for (let attempt = 0; ; attempt++) {
    let response
    try {
      // fetch sets the multipart Content-Type (and its boundary) from the FormData body.
      response = await fetch(url, {
        method,
        headers: { Authorization: `Bearer ${apiKey}` },
        body,
      })
    } catch (error) {
      if (attempt >= delays.length) throw error
      await sleep(delays[attempt])
      continue
    }

    if ((response.status === 429 || response.status >= 500) && attempt < delays.length) {
      await response.body?.cancel()
      await sleep(delays[attempt])
      continue
    }

    const payload = await response.json().catch(() => ({}))
    if (!response.ok) {
      throw new GradeScriptError(
        response.status,
        payload.error ?? response.statusText,
        payload.message ?? ''
      )
    }
    return payload
  }
}

// fields: studentFullname and subject (required), plus any of studentNumber,
//   level, assessmentType and instructions. Any other field is rejected.
// attachments: [{ type, path, index }] where type is SCRIPT, MARKING_GUIDE,
//   QUESTION_PAPER or OTHER, and index (optional, from 1) orders files of the
//   same type. At least one SCRIPT is required.
export async function createGrading(fields, attachments) {
  const form = new FormData()
  for (const [name, value] of Object.entries(fields)) {
    if (value !== undefined) form.append(name, value)
  }
  for (const [i, { type, path, index }] of attachments.entries()) {
    form.append(`attachments[${i}][type]`, type)
    if (index !== undefined) form.append(`attachments[${i}][index]`, String(index))
    form.append(`attachments[${i}][file]`, await openAsBlob(path), basename(path))
  }

  const { data } = await request('POST', '/grading', { body: form })
  return data
}

// Returns { data, pagination }. Pass pagination.nextCursor as cursor for the next page.
export async function listGradings({ status, limit, cursor } = {}) {
  const { data, pagination } = await request('GET', '/grading', {
    query: { status, limit, cursor },
  })
  return { data, pagination }
}

export async function getGrading(id) {
  const { data } = await request('GET', `/grading/${id}`)
  return data
}

export async function retryGrading(id) {
  const { data } = await request('PATCH', `/grading/${id}/retry`)
  return data
}

export async function deleteGrading(id) {
  const { data } = await request('DELETE', `/grading/${id}`)
  return data
}

// Polls a grading every 5 seconds until it finishes. There are no webhooks.
// The server stops a grading after 25 minutes, so the 30-minute timeout is plenty.
export async function waitForGrading(id, { intervalMs = 5000, timeoutMs = 30 * 60 * 1000 } = {}) {
  const deadline = Date.now() + timeoutMs
  let lastStatus

  while (true) {
    const grading = await getGrading(id)
    if (grading.status !== lastStatus) {
      console.log(grading.status)
      lastStatus = grading.status
    }
    if (FINISHED.includes(grading.status)) return grading

    if (Date.now() >= deadline) {
      throw new Error(`Grading ${id} did not finish within ${timeoutMs / 60000} minutes.`)
    }
    await sleep(intervalMs)
  }
}

export function printSummary(grading) {
  console.log(`\nGrading ${grading.id}: ${grading.status}`)
  if (grading.title) console.log(grading.title)
  if (grading.totalMark !== null) {
    console.log(`Mark: ${grading.totalMark} / ${grading.maximumMark}`)
  }
  for (const question of grading.questions) {
    const review = question.needsReview ? ' (needs review)' : ''
    console.log(
      `  ${question.ref}\t${question.awarded} / ${question.max}\t${question.topic}${review}`
    )
  }
  if (grading.feedback) console.log(`\nFeedback:\n${grading.feedback}`)
  if (grading.error) console.log(`\nError: ${grading.error}`)
}

// Runs a command and prints any error as one line instead of a stack trace.
export async function run(main) {
  try {
    await main()
  } catch (error) {
    console.error(error.message)
    process.exitCode = 1
  }
}
