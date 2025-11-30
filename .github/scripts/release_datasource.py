#!/usr/bin/env python3
import os
import sys
import tarfile
import subprocess
from pathlib import Path
from typing import Dict, List

import yaml
import requests


def load_config(config_path: str) -> Dict:
    with open(config_path, "r") as f:
        return yaml.safe_load(f)


def check_git_ref_exists(ref: str) -> bool:
    """Check if a Git ref (tag or branch) exists locally or remotely."""
    try:
        subprocess.run(["git", "rev-parse", "--verify", ref], check=True, capture_output=True)
        return True
    except subprocess.CalledProcessError:
        try:
            result = subprocess.run(
                ["git", "ls-remote", "--refs", "origin", ref],
                check=True,
                capture_output=True,
                text=True
            )
            return bool(result.stdout.strip())
        except subprocess.CalledProcessError:
            return False


class PackageBuilder:
    def __init__(self, config: Dict, datasource: str):
        self.config = config
        self.datasource = datasource
        self.base_path = Path(config["metadata"]["base_path"])
        self.release_config = config["release_config"]
        self.ds_config = config["datasources"][datasource]

        self.package_name = self.ds_config["package_name"]
        self.version = self.ds_config["version"]
        self.dist_dir = Path(self.release_config["distribution_directory"])
        self.dist_dir.mkdir(parents=True, exist_ok=True)

        self.package_files: List[str] = []

    @property
    def archive_name(self) -> str:
        return self.release_config["archive_format"].format(
            package_name=self.package_name, version=self.version
        )

    @property
    def archive_path(self) -> Path:
        return self.dist_dir / self.archive_name

    def build(self) -> Dict[str, str]:
        print(f"📦 Building package for {self.datasource} -> {self.archive_path}")

        with tarfile.open(self.archive_path, "w:gz") as tar:
            for asset in self.config.get("common_assets", []):
                asset_path = self.base_path / asset
                if asset_path.exists():
                    tar.add(asset_path, arcname=asset)
                    self._track_files(asset_path)

            for model_path in self.ds_config.get("model_paths", []):
                path = self.base_path / model_path
                if path.exists():
                    tar.add(path, arcname=model_path)
                    self._track_files(path)

        return {
            "archive": str(self.archive_path),
            "package_name": self.package_name,
            "version": self.version,
        }

    def _track_files(self, path: Path):
        if path.is_file():
            self.package_files.append(str(path))
        else:
            for p in path.rglob("*"):
                if p.is_file():
                    self.package_files.append(str(p))


class GitHubReleaser:
    def __init__(self, token: str, repo_name: str):
        self.token = token
        self.repo = repo_name
        self.api_url = f"https://api.github.com/repos/{repo_name}"

    def release(self, artifact: Dict[str, str], builder: PackageBuilder, branch: str = "main"):
        # Use the exact existing tag
        tag_name = f"ds/{builder.datasource}/v{artifact['version']}"
        archive = artifact["archive"]

        if self._release_exists(tag_name):
            print(f"❌ Release {tag_name} already exists on GitHub.")
            sys.exit(1)

        print(f"🚀 Creating GitHub release {tag_name}")

        release_notes = self._generate_release_notes(artifact, builder.package_files, branch, builder)
        release = self._create_release(tag_name, artifact, release_notes)
        self._upload_asset(release["upload_url"], archive)

        print(f"✅ Uploaded {archive} to release {tag_name}")

    def _headers(self):
        return {"Authorization": f"token {self.token}", "Accept": "application/vnd.github+json"}

    def _release_exists(self, tag_name: str) -> bool:
        url = f"{self.api_url}/releases/tags/{tag_name}"
        r = requests.get(url, headers=self._headers())
        if r.status_code == 200:
            return True
        if r.status_code == 404:
            return False
        r.raise_for_status()

    def _create_release(self, tag_name: str, artifact: Dict[str, str], notes: str) -> Dict:
        url = f"{self.api_url}/releases"
        payload = {
            "tag_name": tag_name,
            "name": f"{artifact['package_name']} v{artifact['version']}",
            "body": notes,
            "draft": False,
            "prerelease": False,
        }
        r = requests.post(url, headers=self._headers(), json=payload)
        r.raise_for_status()
        return r.json()

    def _upload_asset(self, upload_url: str, archive_path: str):
        upload_url = upload_url.split("{")[0]
        filename = Path(archive_path).name
        with open(archive_path, "rb") as f:
            r = requests.post(
                f"{upload_url}?name={filename}",
                headers={
                    "Authorization": f"token {self.token}",
                    "Content-Type": "application/gzip",
                },
                data=f,
            )
        r.raise_for_status()

    def _find_previous_release_tag(self, datasource: str, current_tag: str) -> str:
        """Find the previous tag for this datasource, or None if none exists."""
        try:
            cmd = ["git", "tag", "-l", f"ds/{datasource}/v*", "--sort=-version:refname"]
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            tags = [t.strip() for t in result.stdout.splitlines() if t.strip()]
            for tag in tags:
                if tag != current_tag:
                    return tag
            return None
        except subprocess.CalledProcessError:
            return None

    def _generate_release_notes(
        self, artifact: Dict[str, str], package_files: List[str], branch: str, builder: PackageBuilder
    ) -> str:
        package_name = artifact["package_name"]
        version = artifact["version"]

        base_notes = (
            f"Release {package_name} v{version}\n\n"
            f"Automated release containing packaged DBT models and assets for the **{package_name}** datasource."
        )

        try:
            # Paths relative to repo root
            search_paths = [str(Path(builder.base_path) / p) for p in builder.ds_config.get("model_paths", [])]
            search_paths += [str(Path(builder.base_path) / p) for p in builder.config.get("common_assets", [])]

            current_tag = f"ds/{builder.datasource}/v{version}"
            previous_tag = self._find_previous_release_tag(builder.datasource, current_tag)

            if previous_tag:
                git_range = f"{previous_tag}..{current_tag}"
                cmd = ["git", "log", "--oneline", "--pretty=format:- %s (%h)", git_range, "--", *search_paths]
            else:
                cmd = ["git", "log", "--oneline", "--max-count=10", "--pretty=format:- %s (%h)", "--", *search_paths]

            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            recent_changes = result.stdout.strip()
            notes = f"{base_notes}\n\n## Recent Changes\n{recent_changes}" if recent_changes else base_notes

        except subprocess.CalledProcessError:
            notes = base_notes

        full_history_link = f"https://github.com/{self.repo}/commits/{branch}"
        notes += f"\n\nFor full commit history: {full_history_link}"
        return notes

def main():
    if len(sys.argv) < 3:
        print("Usage: release_datasource.py <datasource> <version>")
        sys.exit(1)

    datasource = sys.argv[1]
    version_from_tag = sys.argv[2].lstrip("v")

    ref_tag = f"refs/tags/ds/{datasource}/v{version_from_tag}"
    if not check_git_ref_exists(ref_tag):
        print(f"❌ Git ref '{ref_tag}' does not exist. Cannot release.")
        sys.exit(1)

    # Resolve config path relative to repo root (where script is executed from in GitHub Actions)
    config_path = Path("datasources-release.yml").resolve()
    config = load_config(str(config_path))

    if datasource not in config.get("datasources", {}):
        print(f"❌ Datasource '{datasource}' not found in config")
        sys.exit(1)

    builder = PackageBuilder(config, datasource)
    builder.version = version_from_tag
    artifact = builder.build()

    token = os.getenv("GITHUB_TOKEN")
    repo_name = os.getenv("GITHUB_REPOSITORY", "ecos-labs/ecos-core")
    branch = os.getenv("GITHUB_REF_NAME", "main")

    if not token or not repo_name:
        print("❌ Missing GITHUB_TOKEN or GITHUB_REPOSITORY environment variables")
        sys.exit(1)

    releaser = GitHubReleaser(token, repo_name)
    releaser.release(artifact, builder, branch)


if __name__ == "__main__":
    main()
