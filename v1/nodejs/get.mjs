// Print one grading in full: marks, per-question results, annotations and attachments.
//
//   node get.mjs <grading-id>
import { getGrading, run } from './gradescript.mjs'

await run(async () => {
  const id = process.argv[2]
  if (!id) throw new Error('Usage: node get.mjs <grading-id>')

  console.log(JSON.stringify(await getGrading(id), null, 2))
})
