// Delete a finished grading (COMPLETED, PARTIAL or FAILED). This can't be undone.
//
//   node delete.mjs <grading-id>
import { deleteGrading, run } from './gradescript.mjs'

await run(async () => {
  const id = process.argv[2]
  if (!id) throw new Error('Usage: node delete.mjs <grading-id>')

  const { id: deletedID } = await deleteGrading(id)
  console.log(`Deleted grading ${deletedID}`)
})
