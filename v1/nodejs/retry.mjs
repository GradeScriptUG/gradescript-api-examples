// Grade a FAILED or PARTIAL grading again, then wait for it to finish.
//
//   node retry.mjs <grading-id>
import { printSummary, retryGrading, run, waitForGrading } from './gradescript.mjs'

await run(async () => {
  const id = process.argv[2]
  if (!id) throw new Error('Usage: node retry.mjs <grading-id>')

  await retryGrading(id)
  console.log(`Re-queued grading ${id}`)

  printSummary(await waitForGrading(id))
})
