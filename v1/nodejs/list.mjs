// List the API gradings your key's owner made in the key's organization, newest first.
//
//   node list.mjs [--status COMPLETED,PARTIAL] [--limit 12] [--cursor <nextCursor>]
import { parseArgs } from 'node:util'

import { listGradings, run } from './gradescript.mjs'

await run(async () => {
  const { values } = parseArgs({
    options: {
      status: { type: 'string' },
      limit: { type: 'string' },
      cursor: { type: 'string' },
    },
  })

  const { data, pagination } = await listGradings(values)

  for (const grading of data) {
    const mark =
      grading.totalMark === null ? '-' : `${grading.totalMark}/${grading.maximumMark}`
    console.log(
      [grading.id, grading.status, grading.studentFullname, mark, grading.createdAt].join('\t')
    )
  }
  if (pagination.hasNextItems) {
    console.log(`\nNext page: node list.mjs --cursor ${pagination.nextCursor}`)
  }
})
