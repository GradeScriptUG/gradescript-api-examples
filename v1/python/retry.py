"""Grade a FAILED or PARTIAL grading again, then wait for it to finish.

    python retry.py <grading-id>
"""
import argparse

from gradescript import print_summary, retry_grading, run, wait_for_grading


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("grading_id")
    args = parser.parse_args()

    retry_grading(args.grading_id)
    print(f"Re-queued grading {args.grading_id}", flush=True)

    print_summary(wait_for_grading(args.grading_id))


if __name__ == "__main__":
    run(main)
