"""Poll a grading until it finishes, then print its marks and feedback.

    python wait.py <grading-id>
"""
import argparse

from gradescript import print_summary, run, wait_for_grading


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("grading_id")
    args = parser.parse_args()

    print_summary(wait_for_grading(args.grading_id))


if __name__ == "__main__":
    run(main)
