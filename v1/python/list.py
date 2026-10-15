"""List the API gradings your key's owner made in the key's organization, newest first.

    python list.py [--status COMPLETED,PARTIAL] [--limit 12] [--cursor <nextCursor>]
"""
import argparse

from gradescript import list_gradings, run


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--status", help="comma-separated, e.g. COMPLETED,PARTIAL")
    parser.add_argument("--limit", type=int, help="1 to 25, default 12")
    parser.add_argument("--cursor", help="nextCursor from the previous page")
    args = parser.parse_args()

    page = list_gradings(status=args.status, limit=args.limit, cursor=args.cursor)

    for grading in page["data"]:
        mark = (
            "-"
            if grading["totalMark"] is None
            else f"{grading['totalMark']:g}/{grading['maximumMark']:g}"
        )
        print(
            "\t".join(
                [
                    grading["id"],
                    grading["status"],
                    grading["studentFullname"],
                    mark,
                    grading["createdAt"],
                ]
            )
        )
    if page["pagination"]["hasNextItems"]:
        print(f"\nNext page: python list.py --cursor {page['pagination']['nextCursor']}")


if __name__ == "__main__":
    run(main)
