import pytest
from go_archiver import Archiver

def test_create_archiver():
    archiver = Archiver(
        sourcePath="src/go_archiver/go/archiver/testdata/source",
        outputPath="test_output.tar.gz",
        recursive=True,
        filterMode="all"
    )
    result = archiver.Archive()
    assert result == ""  # Empty string means success

if __name__ == '__main__':
    pytest.main() 