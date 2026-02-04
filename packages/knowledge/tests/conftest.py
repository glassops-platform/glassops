import sys
from pathlib import Path

# Add 'packages' to sys.path so 'knowledge' package can be imported
# This file is in packages/knowledge/tests/
# We want to add packages/ (which is ../../ from here)
TEST_DIR = Path(__file__).resolve().parent
PACKAGES_DIR = TEST_DIR.parent.parent

# This needs to be absolute
if str(PACKAGES_DIR) not in sys.path:
    print(f"Adding to sys.path: {PACKAGES_DIR}")
    sys.path.insert(0, str(PACKAGES_DIR))
