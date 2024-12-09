import typer
from typing import Optional
from pathlib import Path
from .cli.prompt import get_path_input, show_spinner, show_progress
from .cmd import tarball

app = typer.Typer(
    name="go-archiver",
    help="A CLI tool for creating and managing tarballs with Go-powered compression",
    add_completion=True,
)

# Register commands from cmd module
app.add_typer(tarball.app, name="tarball")

@app.callback()
def callback():
    """
    Go-Archiver: Efficient tarball management with Go-powered compression
    """
    pass

def main():
    app()

if __name__ == "__main__":
    main()
