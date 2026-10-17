# Node.js examples

ES modules that call the GradeScript Grading API with the built-in `fetch`, `FormData` and
`openAsBlob`, and read the `.env` file with `process.loadEnvFile`. There's nothing to install.

## Requirements

- Node.js 20.12 or later
- `GRADESCRIPT_API_KEY` in the `.env` file at the repo root (see the
  [main README](../../README.md#before-you-start))

## Usage

```bash
node grade.mjs                            # grade the sample script and wait for the result
node grade.mjs page-1.jpg page-2.jpg      # each file is one SCRIPT page, in order
node wait.mjs <grading-id>                # wait for a grading and print its marks
node list.mjs --status COMPLETED,PARTIAL  # list gradings; also --limit N and --cursor C
node get.mjs <grading-id>                 # print the full grading as JSON
node retry.mjs <grading-id>               # grade a FAILED or PARTIAL grading again
node delete.mjs <grading-id>              # delete a finished grading
```

The same commands work through npm, for example `npm run grade` or
`npm run list -- --status COMPLETED`.

## Using the client in your own code

[`gradescript.mjs`](gradescript.mjs) is a small client you can copy into your project:

```js
import { createGrading, waitForGrading } from './gradescript.mjs'

const queued = await createGrading(
  { studentFullname: 'Ada Lovelace', subject: 'Mathematics', level: 'S.2' },
  [
    { type: 'SCRIPT', path: 'page-1.jpg', index: 1 },
    { type: 'SCRIPT', path: 'page-2.jpg', index: 2 },
    { type: 'MARKING_GUIDE', path: 'marking-guide.pdf' },
  ]
)

const grading = await waitForGrading(queued.id)
console.log(grading.totalMark, '/', grading.maximumMark)
```

It also exports `listGradings`, `getGrading`, `retryGrading` and `deleteGrading`. API errors are
thrown as a `GradeScriptError`, which has `statusCode` and `error` (for example `PAYMENT_REQUIRED`).
GET requests are retried on network errors, 429 and 5xx. Create and retry requests aren't, so a
script is never graded twice.
