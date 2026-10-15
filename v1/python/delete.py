"""Delete a finished grading (COMPLETED, PARTIAL or FAILED). This can't be undone.

    python delete.py <grading-id>
"""
import argparse

from gradescript import delete_grading, run


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("grading_id")
    args = parser.parse_args()

    deleted = delete_grading(args.grading_id)
    print(f"Deleted grading {deleted['id']}")


if __name__ == "__main__":
    run(main)
