// Poll a grading until it finishes, then print its marks and feedback.
//
//   node wait.mjs <grading-id>
import { printSummary, run, waitForGrading } from './gradescript.mjs'

await run(async () => {
  const id = process.argv[2]
  if (!id) throw new Error('Usage: node wait.mjs <grading-id>')

  printSummary(await waitForGrading(id))
})
