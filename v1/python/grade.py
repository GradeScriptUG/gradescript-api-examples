"""Grade a script: upload it, then wait for the grading to finish.

    python grade.py                        grades the sample script in ../../sample-script
    python grade.py page-1.jpg page-2.jpg  each file is one SCRIPT page, in order
"""
import argparse
import os

from gradescript import create_grading, print_summary, run, wait_for_grading

SAMPLE_SCRIPT = os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "..",
    "..",
    "sample-script",
    "Item-one-answer-script.jpeg",
)


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("files", nargs="*", default=[SAMPLE_SCRIPT], help="script pages, in order")
    args = parser.parse_args()

    grading = create_grading(
        {"studentFullname": "Ada Lovelace", "subject": "Mathematics", "level": "S.2"},
        [
            {"type": "SCRIPT", "path": path, "index": i + 1}
            for i, path in enumerate(args.files)
        ],
    )
    print(f"Queued grading {grading['id']}", flush=True)

    print_summary(wait_for_grading(grading["id"]))


if __name__ == "__main__":
    run(main)
