import os
import shutil
import subprocess
import sys
import re
import argparse
from pathlib import Path

def main():
    parser = argparse.ArgumentParser(description="Build or Serve GlassOps documentation")
    parser.add_argument("--serve", action="store_true", help="Serve the documentation after staging")
    parser.add_argument("--dirty", action="store_true", help="Fast rebuilds (for serve mode)")
    args = parser.parse_args()

    # Define paths
    root_dir = Path(__file__).parent.parent
    config_dir = root_dir / "config"
    staging_dir = root_dir / "docs_staging"
    output_dir = root_dir / "glassops_site"

    # Helper to handle read-only files on Windows
    def on_rm_error(func, path, exc_info):
        import stat
        os.chmod(path, stat.S_IWRITE)
        os.unlink(path)

    # Clean previous build
    if staging_dir.exists():
        print("Cleaning staging directory...")
        shutil.rmtree(staging_dir, onerror=on_rm_error)
    
    # Only clean output if we are rebuilding
    if not args.serve and output_dir.exists():
        print("Cleaning output directory...")
        shutil.rmtree(output_dir, onerror=on_rm_error)
    
    staging_dir.mkdir()
    content_dir = staging_dir / "content"
    content_dir.mkdir()

    # Define what to copy (allow-list approach)
    copy_list = [
        "docs",
        "packages",
        "examples",
        "CONTRIBUTING.md",
        "README.md",
        "LICENSE"
    ]

    print(f"Staging documentation from {root_dir} to {content_dir}...")

    # Exclude patterns
    # We use a set for faster lookups
    EXCLUDED_DIRS = {
        "node_modules", ".git", ".DS_Store", "dist", "build", "coverage", 
        ".turbo", "__pycache__", "venv", ".venv", ".pytest_cache"
    }
    EXCLUDED_EXTENSIONS = {".map", ".pyc"}

    def ignore_patterns(path, names):
        ignored = []
        for name in names:
            if name in EXCLUDED_DIRS:
                ignored.append(name)
            elif any(name.endswith(ext) for ext in EXCLUDED_EXTENSIONS):
                ignored.append(name)
        return ignored

    for item in copy_list:
        src = root_dir / item
        dst = content_dir / item
        if src.exists():
            if src.is_dir():
                shutil.copytree(src, dst, ignore=ignore_patterns)
            else:
                shutil.copy2(src, dst)
        else:
            print(f"Warning: Source item {item} not found")

    # Copy config to root of staging (outside content)
    shutil.copy2(config_dir / "mkdocs.yml", staging_dir / "mkdocs.yml")

    # Modify mkdocs.yml in staging
    mkdocs_config_path = staging_dir / "mkdocs.yml"
    with open(mkdocs_config_path, "r") as f:
        config_content = f.read()

    # Fix paths in config
    lines = config_content.splitlines()
    new_lines = []
    for line in lines:
        if line.strip().startswith("docs_dir:"):
            new_lines.append("docs_dir: content")
        elif line.strip().startswith("site_dir:"):
            new_lines.append(f"site_dir: ../glassops_site")
        else:
            new_lines.append(line)
    
    with open(mkdocs_config_path, "w") as f:
        f.write("\n".join(new_lines))

    # Post-processing: Fix conflicts and links
    print("Post-processing staged files...")
    
    # Generate index for docs/adr if it doesn't exist
    adr_dir = content_dir / "docs" / "adr"
    if adr_dir.exists() and not (adr_dir / "README.md").exists() and not (adr_dir / "index.md").exists():
        print("Generating index for docs/adr...")
        with open(adr_dir / "README.md", "w") as f:
            f.write("# Protocol ADRs\n\n")
            f.write("Index of Architectural Decision Records:\n\n")
            for adr_file in sorted(adr_dir.glob("*.md")):
                 f.write(f"- [{adr_file.stem}]({adr_file.name})\n")

    # Rename LICENSE files to LICENSE.md so they render in MkDocs
    print("Renaming LICENSE files to LICENSE.md...")
    for license_file in content_dir.rglob("LICENSE"):
        if license_file.is_file():
            new_path = license_file.with_suffix(".md")
            # Read content, wrap in code block or preserve? 
            # LICENSE files are usually text. Let's wrap them in a pre block or just rename.
            # If we just rename, they render as text which is fine.
            # Let's just rename for now.
            license_file.rename(new_path)

    for file_path in content_dir.rglob("*.md"):
        # 1. Handle README.md vs index.md collisions
        if file_path.name == "README.md":
            index_path = file_path.parent / "index.md"
            if index_path.exists():
                print(f"Resolving collision: Renaming {file_path.name} in {file_path.parent.name} to overview.md")
                new_path = file_path.parent / "overview.md"
                file_path.rename(new_path)
                file_path = new_path # Update reference for link fixing

        # 2. Fix Links
        try:
            content = file_path.read_text(encoding="utf-8")
            original_content = content

            # 2.1 Convert GitHub Alerts (> [!NOTE]) to MkDocs Admonitions (!!! note)
            lines = content.splitlines()
            new_lines = []
            in_alert = False
            # GitHub Alert -> Material Admonition mapping
            ALERT_MAP = {
                "note": "note",
                "tip": "tip",
                "important": "important",
                "warning": "warning",
                "caution": "danger"
            }
            for line in lines:
                # Support: "> [!NOTE]" or ">[!NOTE]" or "  > [!NOTE] content"
                alert_match = re.match(r'^\s*>\s*\[!(\w+)\]\s*(.*)', line)
                if alert_match:
                    in_alert = True
                    raw_type = alert_match.group(1).lower()
                    alert_type = ALERT_MAP.get(raw_type, raw_type)
                    content_after = alert_match.group(2).strip()
                    new_lines.append(f"!!! {alert_type}")
                    if content_after:
                        new_lines.append(f"    {content_after}")
                elif in_alert and (line.strip().startswith(">") or line.strip() == ""):
                    # Clean up the "> " prefix if it exists
                    content_line = re.sub(r'^\s*>\s?', '', line)
                    new_lines.append(f"    {content_line}")
                else:
                    in_alert = False
                    new_lines.append(line)
            content = "\n".join(new_lines)
            
            # 2.2 Using Regex to specifically target directory links
            # Matches: ]( ... docs/adr ) or ]( ... docs/adr/ )
            # Replacement: ]( ... docs/adr/README.md )
            content = re.sub(r'\]\(([^)]*docs/adr/?)\)', r'](\1/README.md)', content)
            
            # Cleanup double READMEs if they happened
            content = content.replace("/README.md/README.md", "/README.md")
            
            # Specific Fixes ONLY (No blind replace)

            # Fix links to LICENSE files which were renamed to LICENSE.md
            content = re.sub(r'\]\(([^)]*LICENSE)\)', r'](\1.md)', content)
            
            # Fix links to root README which was renamed to overview.md (legacy logic)
            # Actually, we now use README.md as the home page and deleted index.md
            # So we fix links pointing to index.md to point to README.md
            content = content.replace("/index.md", "/README.md")
            # Handle cases where index.md is linked relatively from root
            if "docs" in file_path.parts:
                 content = content.replace("](../README.md)", "](../README.md)") # No change needed here
            
            if content != original_content:
                file_path.write_text(content, encoding="utf-8")
                
        except Exception as e:
            print(f"Warning processing {file_path}: {e}")

    # Build or Serve
    if args.serve:
        print("Serving MkDocs site...")
        cmd = [sys.executable, "-m", "mkdocs", "serve"]
        if args.dirty:
            cmd.append("--dirty")
    else:
        print("Building MkDocs site...")
        cmd = [sys.executable, "-m", "mkdocs", "build"]
    
    try:
        # We don't capture output in serve mode so user can see logs
        subprocess.run(
            cmd,
            cwd=staging_dir,
            check=True
        )
        if not args.serve:
            print(f"Build successful! Output at: {output_dir}")
            
    except subprocess.CalledProcessError as e:
        print(f"Command failed with exit code {e.returncode}")
        sys.exit(e.returncode)

if __name__ == "__main__":
    main()
