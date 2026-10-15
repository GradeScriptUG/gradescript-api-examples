"""Print one grading in full: marks, per-question results, annotations and attachments.

    python get.py <grading-id>
"""
import argparse
import json

from gradescript import get_grading, run


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("grading_id")
    args = parser.parse_args()

    print(json.dumps(get_grading(args.grading_id), indent=2, ensure_ascii=False))


if __name__ == "__main__":
    run(main)
