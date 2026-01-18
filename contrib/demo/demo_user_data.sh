#! /bin/bash


export DEPLOY_SECRET_KEY="demo-secret-key"
export DEPLOY_PORT=8888
export WORK_DIR=$(pwd)/demo_data
export STATE_FILE="$WORK_DIR/.deploy_state.json"
export POTATO_PORT=80
export POTATO_HOST="*.tubersalltheway.top"

mkdir -p $WORK_DIR

cd $WORK_DIR

PYTHON_CODE=$(cat <<EOF
#!/usr/bin/env python3

"""
Demo deployment server for potatoverse

GET /deploy/:secret_key
- Downloads latest tar.gz file from GitHub releases
- Extracts tar.gz file
- If .pdata/maindb doesn't exist, runs potatoverse server init
- Runs ./potatoverse server start

POST /reset/:secret_key
- Deletes .pdata/maindb (not config)
"""

import os
import sys
import json
import shutil
import tarfile
import subprocess
import platform
import tempfile
import signal
import time
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs
from urllib.request import urlopen, Request
from urllib.error import URLError


# Configuration
GITHUB_REPO = "blue-monads/potatoverse"
SECRET_KEY = os.environ.get("DEPLOY_SECRET_KEY", "demo-secret-key")
PORT = int(os.environ.get("DEPLOY_PORT", "8888"))
WORK_DIR = os.environ.get("WORK_DIR", os.getcwd())
# Ensure WORK_DIR is an absolute path
WORK_DIR = os.path.abspath(WORK_DIR)
STATE_FILE = os.path.join(WORK_DIR, ".deploy_state.json")
POTATO_PORT = int(os.environ.get("POTATO_PORT", "80"))
POTATO_HOST = os.environ.get("POTATO_HOST", "*.tubersalltheway.top")

# Change to working directory early
os.makedirs(WORK_DIR, exist_ok=True)
os.chdir(WORK_DIR)


def get_latest_release_info():
    """Get the latest release info including version and asset URL from GitHub."""
    api_url = f"https://api.github.com/repos/{GITHUB_REPO}/releases/latest"
    
    try:
        req = Request(api_url)
        req.add_header("Accept", "application/vnd.github.v3+json")
        
        with urlopen(req) as response:
            data = json.loads(response.read())
            
        version = data.get("tag_name", "").lstrip("v")  # Remove 'v' prefix if present
        if not version:
            version = data.get("name", "unknown")
        
        # Find tar.gz asset for current platform
        system = platform.system().lower()
        machine = platform.machine().lower()
        
        # Map machine names
        if machine in ["x86_64", "amd64"]:
            arch = "amd64"
        elif machine in ["aarch64", "arm64"]:
            arch = "arm64"
        else:
            arch = machine
        
        # Map system names to release asset naming
        os_name_map = {
            "linux": "linux",
            "darwin": "darwin",
            "windows": "windows"
        }
        os_name = os_name_map.get(system, system)
        
        print(f"Looking for asset matching OS: {os_name}, Arch: {arch}")
        
        # Look for matching asset - must match both OS and arch
        # Format is typically: potatoverse_VERSION_OS_ARCH.tar.gz
        for asset in data.get("assets", []):
            name = asset["name"]
            if name.endswith(".tar.gz"):
                name_lower = name.lower()
                # Check if it matches our platform exactly
                # The OS and arch should both be in the filename
                if os_name in name_lower and f"_{arch}" in name_lower:
                    # Verify the OS name appears in the filename (not as part of another word)
                    # and that we're not matching a different OS
                    # Pattern: _os_arch or _os_arch.tar.gz
                    pattern = f"_{os_name}_{arch}"
                    if pattern in name_lower:
                        print(f"Found matching asset: {name}")
                        return version, asset["browser_download_url"]
        
        # Fallback: return first tar.gz if no match
        for asset in data.get("assets", []):
            if asset["name"].endswith(".tar.gz"):
                return version, asset["browser_download_url"]
        
        raise ValueError("No tar.gz asset found in latest release")
        
    except URLError as e:
        raise Exception(f"Failed to fetch release info: {e}")


def get_latest_release_asset():
    """Get the latest release tar.gz asset URL from GitHub (backward compatibility)."""
    _, url = get_latest_release_info()
    return url


def download_file(url, dest_path):
    """Download a file from URL to destination path."""
    print(f"Downloading {url}...")
    req = Request(url)
    req.add_header("Accept", "application/octet-stream")
    
    with urlopen(req) as response:
        with open(dest_path, "wb") as f:
            shutil.copyfileobj(response, f)
    
    print(f"Downloaded to {dest_path}")


def extract_tar_gz(tar_path, extract_to):
    """Extract tar.gz file to directory."""
    print(f"Extracting {tar_path} to {extract_to}...")
    os.makedirs(extract_to, exist_ok=True)
    
    with tarfile.open(tar_path, "r:gz") as tar:
        tar.extractall(extract_to)
    
    print(f"Extracted to {extract_to}")


def find_potatoverse_binary(extract_dir):
    """Find the potatoverse binary in extracted directory."""
    for root, dirs, files in os.walk(extract_dir):
        for file in files:
            if file == "potatoverse" or (platform.system() == "Windows" and file == "potatoverse.exe"):
                return os.path.join(root, file)
    return None


def load_state():
    """Load deployment state from file."""
    if os.path.exists(STATE_FILE):
        try:
            with open(STATE_FILE, "r") as f:
                return json.load(f)
        except (json.JSONDecodeError, IOError):
            return {"version": None, "pid": None}
    return {"version": None, "pid": None}


def save_state(version, pid):
    """Save deployment state to file."""
    state = {"version": version, "pid": pid}
    with open(STATE_FILE, "w") as f:
        json.dump(state, f)


def is_process_running(pid):
    """Check if a process with given PID is running."""
    if pid is None:
        return False
    try:
        # Check if process exists (doesn't kill it, just checks)
        os.kill(pid, 0)
        return True
    except (OSError, ProcessLookupError):
        return False


def find_server_pid():
    """Find the actual server process PID (child process listening on port)."""
    # Try to find process listening on port 80 (or check for potatoverse processes)
    try:
        # Use lsof to find process listening on port 80
        result = subprocess.run(
            ["lsof", "-ti:80"],
            capture_output=True,
            text=True,
            timeout=2
        )
        if result.returncode == 0 and result.stdout.strip():
            pid = int(result.stdout.strip().split()[0])
            return pid
    except (subprocess.TimeoutExpired, ValueError, FileNotFoundError, IndexError):
        pass
    
    # Fallback: find potatoverse server actual-start process
    try:
        result = subprocess.run(
            ["pgrep", "-f", "potatoverse server actual-start"],
            capture_output=True,
            text=True,
            timeout=2
        )
        if result.returncode == 0 and result.stdout.strip():
            pid = int(result.stdout.strip().split()[0])
            return pid
    except (subprocess.TimeoutExpired, ValueError, FileNotFoundError, IndexError):
        pass
    
    return None


def kill_server_process(pid):
    """Kill the server process and its children."""
    if not pid or not is_process_running(pid):
        return
    
    print(f"Killing server process {pid} and its children...")
    try:
        # Kill all child processes first
        try:
            subprocess.run(["pkill", "-P", str(pid)], timeout=5)
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass
        
        # Kill the main process
        try:
            os.kill(pid, signal.SIGTERM)
            # Wait a bit for graceful shutdown
            time.sleep(2)
            # Force kill if still running
            if is_process_running(pid):
                os.kill(pid, signal.SIGKILL)
        except (OSError, ProcessLookupError):
            pass
        
        print(f"Server process {pid} killed")
    except Exception as e:
        print(f"Error killing process {pid}: {e}")


def deploy():
    """Deploy the latest release."""
    print("Starting deployment...")
    print(f"Working directory: {WORK_DIR}")
    print(f"Current directory: {os.getcwd()}")
    
    # Load current state
    state = load_state()
    current_version = state.get("version")
    current_pid = state.get("pid")
    
    # Get latest release info
    latest_version, asset_url = get_latest_release_info()
    print(f"Latest version: {latest_version}, Current version: {current_version}")
    
    # Check if latest version is already running
    if current_version == latest_version and current_pid:
        if is_process_running(current_pid):
            print(f"Version {latest_version} is already running (PID: {current_pid})")
            return {
                "status": "success",
                "message": f"Version {latest_version} is already running",
                "version": latest_version,
                "pid": current_pid
            }
        else:
            print(f"Version {latest_version} was running but process {current_pid} is not alive")
    
    # If different version or process not running, proceed with deployment
    if current_version != latest_version and current_pid:
        print(f"New version detected ({current_version} -> {latest_version}), killing old process...")
        kill_server_process(current_pid)
        # Also try to find and kill any orphaned processes
        found_pid = find_server_pid()
        if found_pid and found_pid != current_pid:
            kill_server_process(found_pid)
    
    # Create temporary directory for download
    with tempfile.TemporaryDirectory() as tmpdir:
        target_binary = os.path.join(WORK_DIR, "potatoverse")
        if platform.system() == "Windows":
            target_binary += ".exe"
        
        if current_version != latest_version:
            tar_path = os.path.join(tmpdir, "potatoverse.tar.gz")

            # Download latest release
            download_file(asset_url, tar_path)
            
            # Extract
            extract_dir = os.path.join(tmpdir, "extracted")
            extract_tar_gz(tar_path, extract_dir)
            
            # Find binary
            binary_path = find_potatoverse_binary(extract_dir)
            if not binary_path:
                raise Exception("potatoverse binary not found in release")
            
            # Copy binary to work directory

            
            shutil.copy2(binary_path, target_binary)
            os.chmod(target_binary, 0o755)
            print(f"Copied binary to {target_binary}")        

        
        # Check if .pdata/maindb exists
        maindb_path = os.path.join(WORK_DIR, ".pdata", "maindb")
        needs_init = not os.path.exists(maindb_path)
        
        if needs_init:
            print("Initializing server...")
            result = subprocess.run(
                [target_binary, "server", "init", "--port", str(POTATO_PORT), "--host", POTATO_HOST],
                capture_output=True,
                text=True
            )
            if result.returncode != 0:
                raise Exception(f"Server init failed: {result.stderr}")
            print("Server initialized")
        
        # Start server in background
        print("Starting server...")
        process = subprocess.Popen(
            [target_binary, "server", "start", "--auto-seed"],
            stdout=None,  # Share stdout with parent
            stderr=None,  # Share stderr with parent
            cwd=WORK_DIR
        )
        
        # Wait a moment for the server to start and find the actual server PID
        time.sleep(3)
        server_pid = find_server_pid()
        
        if server_pid:
            print(f"Server started, PID: {server_pid} (parent: {process.pid})")
            save_state(latest_version, server_pid)
        else:
            print(f"Server started (parent PID: {process.pid}), but couldn't find actual server process")
            # Save parent PID as fallback
            save_state(latest_version, process.pid)
    
    return {
        "status": "success",
        "message": f"Deployment completed, version {latest_version}",
        "version": latest_version,
        "pid": server_pid if 'server_pid' in locals() else process.pid
    }


def get_status():
    """Get current deployment status."""
    state = load_state()
    version = state.get("version")
    pid = state.get("pid")
    
    running = False
    if pid:
        running = is_process_running(pid)
        # Also check if we can find the server process
        if not running:
            found_pid = find_server_pid()
            if found_pid:
                running = True
                pid = found_pid
                # Update state with found PID
                save_state(version, pid)
    
    # Get latest version info
    try:
        latest_version, _ = get_latest_release_info()
        is_latest = (version == latest_version) if version else False
    except Exception as e:
        latest_version = None
        is_latest = None
    
    return {
        "version": version,
        "latest_version": latest_version,
        "is_latest": is_latest,
        "pid": pid,
        "running": running
    }


def reset():
    """Reset the demo by deleting .pdata/maindb."""
    print("Resetting demo state...")
    
    # Note: We don't kill the server here, just reset the database
    # The server will continue running with a fresh database
    
    maindb_path = os.path.join(WORK_DIR, ".pdata", "maindb")
    
    if os.path.exists(maindb_path):
        if os.path.isdir(maindb_path):
            shutil.rmtree(maindb_path)
        else:
            os.remove(maindb_path)
        print(f"Deleted {maindb_path}")
        return {"status": "success", "message": "Demo state reset"}
    else:
        print(f"{maindb_path} does not exist")
        return {"status": "success", "message": "Demo state already clean"}


class DeployHandler(BaseHTTPRequestHandler):
    """HTTP request handler for deployment endpoints."""

    def do_GET(self):
        parsed_path = urlparse(self.path)
        path_parts = parsed_path.path.strip("/").split("/")

        if len(path_parts) == 2 and path_parts[0] == "status":
            secret_key = path_parts[1]
            
            if secret_key != SECRET_KEY:
                self.send_response(401)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Unauthorized"}).encode())
                return
            
            try:
                result = get_status()
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(result).encode())
            except Exception as e:
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
        else:
            self.send_response(404)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())
    
    def perform_deploy(self):
        """Handle deploy requests."""
        parsed_path = urlparse(self.path)
        path_parts = parsed_path.path.strip("/").split("/")
        
        if len(path_parts) == 2 and path_parts[0] == "deploy":
            secret_key = path_parts[1]
            
            if secret_key != SECRET_KEY:
                self.send_response(401)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Unauthorized"}).encode())
                return
            
            try:
                result = deploy()
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(result).encode())
            except Exception as e:
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())

    
    def do_POST(self):
        """Handle POST requests."""
        parsed_path = urlparse(self.path)
        path_parts = parsed_path.path.strip("/").split("/")

        if len(path_parts) == 2 and path_parts[0] == "deploy":
            self.perform_deploy()
            return

        
        if len(path_parts) == 2 and path_parts[0] == "reset":
            secret_key = path_parts[1]
            
            if secret_key != SECRET_KEY:
                self.send_response(401)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Unauthorized"}).encode())
                return
            
            try:
                result = reset()
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(result).encode())
            except Exception as e:
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
        else:
            self.send_response(404)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())
    
    def log_message(self, format, *args):
        """Override to use print instead of stderr."""
        print(f"[{self.address_string()}] {format % args}")


def main():
    """Main entry point."""
    print(f"Starting deployment server on port {PORT}")
    print(f"Working directory: {WORK_DIR}")
    print(f"Current directory: {os.getcwd()}")
    print(f"Secret key: {SECRET_KEY[:10]}...")
    print(f"GitHub repo: {GITHUB_REPO}")
    
    # Ensure we're in the working directory
    os.chdir(WORK_DIR)
    print(f"Changed to working directory: {os.getcwd()}")
    
    server = HTTPServer(("", PORT), DeployHandler)
    
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down server...")
        server.shutdown()


if __name__ == "__main__":
    main()




EOF
)

echo "$PYTHON_CODE" | python3