import os
import subprocess
from pathlib import Path
from hatchling.builders.hooks.plugin.interface import BuildHookInterface

class CustomBuildHook(BuildHookInterface):
    def initialize(self, version, build_data):
        """Initialize build hook."""
        self._build_go()
        return super().initialize(version, build_data)

    def _build_go(self):
        """Build Go bindings using gopy."""
        try:
            # Get the directory containing this script
            current_dir = Path(__file__).parent
            go_dir = current_dir / "go"
            
            # Store current directory
            original_dir = os.getcwd()
            
            try:
                # Change to go directory and run make
                os.chdir(go_dir)
                result = subprocess.run(
                    ["make", "build-bindings"], 
                    check=True, 
                    capture_output=True, 
                    text=True
                )
                print(result.stdout)
            finally:
                # Always return to original directory
                os.chdir(original_dir)
            
        except subprocess.CalledProcessError as e:
            print(f"Error building Go bindings: {e}")
            print(f"stdout: {e.stdout}")
            print(f"stderr: {e.stderr}")
            raise RuntimeError("Failed to build Go bindings") from e
        except Exception as e:
            print(f"Unexpected error: {e}")
            raise RuntimeError("Failed to build Go bindings") from e


def build():
    """CLI entry point for manual builds."""
    hook = CustomBuildHook(str(Path(__file__).parent), {}, None, None, "", "", None)
    hook._build_go()

if __name__ == "__main__":
    build() 