from prompt_toolkit import PromptSession
from prompt_toolkit.completion import PathCompleter
from prompt_toolkit.shortcuts import ProgressBar
from prompt_toolkit.formatted_text import HTML
from prompt_toolkit.shortcuts import create_progress_bar
from prompt_toolkit.shortcuts.progress_bar import formatters
from prompt_toolkit.formatted_text import HTML
from typing import Optional
import time
from contextlib import contextmanager

def get_path_input(prompt_text: str, default: Optional[str] = None) -> str:
    """Get path input with autocomplete functionality"""
    session = PromptSession()
    completer = PathCompleter()
    
    result = session.prompt(
        HTML(f"<ansiblue>{prompt_text}</ansiblue> "),
        completer=completer,
        default=default or ""
    )
    
    return result

@contextmanager
def show_spinner(text: str):
    """Show a spinner while executing a task"""
    title = HTML(f"<b>{text}</b>")
    formatters_list = [
        formatters.SpinnerFormatter(text=title),
    ]
    
    with create_progress_bar(formatters=formatters_list) as pb:
        task = pb.add_task(text)
        try:
            yield
        finally:
            pb.remove_task(task)

def show_progress(total: int, description: str = "Processing"):
    """Show a progress bar for long-running operations"""
    with ProgressBar() as pb:
        for i in pb(range(total), label=description):
            yield i
