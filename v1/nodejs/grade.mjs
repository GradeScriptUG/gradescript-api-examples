// Grade a script: upload it, then wait for the grading to finish.
//
//   node grade.mjs                        grades the sample script in ../../sample-script
//   node grade.mjs page-1.jpg page-2.jpg  each file is one SCRIPT page, in order
import { fileURLToPath } from 'node:url'

import { createGrading, printSummary, run, waitForGrading } from './gradescript.mjs'

const sampleScript = fileURLToPath(
  new URL('../../sample-script/Item-one-answer-script.jpeg', import.meta.url)
)

await run(async () => {
  const files = process.argv.slice(2)
  if (files.length === 0) files.push(sampleScript)

  const grading = await createGrading(
    { studentFullname: 'Ada Lovelace', subject: 'Mathematics', level: 'S.2' },
    files.map((path, i) => ({ type: 'SCRIPT', path, index: i + 1 }))
  )
  console.log(`Queued grading ${grading.id}`)

  printSummary(await waitForGrading(grading.id))
})
