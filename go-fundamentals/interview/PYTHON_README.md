# Python Solution - Quick Reference

This Python implementation **mirrors the Go solution** structure and approach.

## Quick Start

```bash
# Run with default input.txt
python3 solution.py

# Run with custom input file
python3 solution.py --inputFile input.txt

# Run tests
python3 -m pytest test_solution.py -v
```

## Output Format

```
[DirInfo(path='user', size=31673), DirInfo(path='home', size=14880187), DirInfo(path='log', size=584), DirInfo(path='var', size=584), DirInfo(path='/', size=14880771)]
32841
```

- **Line 1**: List of all directories with their sizes
- **Line 2**: Sum of directories under 100,000 bytes (user: 31,673 + log: 584 + var: 584 = 32,841)

## File Structure

```
interview/
├── solution.go              # Original Go implementation
├── solution.py              # Python implementation (mirrors Go)
├── test_solution.py         # Python test suite (16 tests)
├── input.txt                # Sample input data
├── README.md                # Problem description
├── PYTHON_SOLUTION.md       # Detailed Python documentation
├── GO_PYTHON_COMPARISON.md  # Side-by-side comparison
└── PYTHON_README.md         # This file
```

## Function Overview

| Python Function | Go Equivalent | Purpose |
|---|---|---|
| `new_node()` | `NewNode()` | Create node instances |
| `build_file_system()` | `buildFileSystem()` | Parse input and build tree |
| `calculate_size()` | `calculateSize()` | Recursive size calculation |
| `find_dirs_under_limit()` | `findDirsUnderLimit()` | Find dirs below threshold |
| `find_dirs_size()` | `findDirsSize()` | Get all directory info |
| `run()` | `Run()` | Main execution |

## Key Features

### ✅ Mirrors Go Implementation
- Standalone functions (not class methods)
- Same algorithmic approach
- Parallel naming conventions
- Identical output format

### ✅ Comprehensive Testing
- 16 unit tests covering all functions
- Integration tests with example data
- 100% test pass rate

### ✅ Type Safety
- Type hints throughout
- Dataclass for `DirInfo`
- Optional types for nullable fields

### ✅ Command-Line Interface
- Argparse for CLI arguments
- Matches Go's `--inputFile` flag
- Defaults to `input.txt`

## Example Usage

### Basic Run
```bash
$ python3 solution.py
[DirInfo(path='user', size=31673), DirInfo(path='home', size=14880187), DirInfo(path='log', size=584), DirInfo(path='var', size=584), DirInfo(path='/', size=14880771)]
32841
```

### Run Tests
```bash
$ python3 -m pytest test_solution.py -v
============================= test session starts ==============================
collected 16 items

test_solution.py::TestNode::test_new_node_directory PASSED               [  6%]
test_solution.py::TestNode::test_new_node_file PASSED                    [ 12%]
...
============================== 16 passed in 0.01s ===============================
```

## Implementation Highlights

### Node Structure
```python
class Node:
    name: str
    is_dir: bool
    size: int
    children: Dict[str, Node]
    parent: Optional[Node]
```

### Constructor Pattern
```python
def new_node(name: str, is_dir: bool, size: int) -> Node:
    return Node(name, is_dir, size)
```

### Recursive Size Calculation
```python
def calculate_size(node: Node) -> int:
    if not node.is_dir:
        return node.size
    return sum(calculate_size(child) for child in node.children.values())
```

### Directory Info
```python
@dataclass
class DirInfo:
    path: str
    size: int
```

## Documentation

- **`PYTHON_SOLUTION.md`** - Complete documentation with examples and design decisions
- **`GO_PYTHON_COMPARISON.md`** - Detailed side-by-side comparison of Go and Python implementations
- **`README.md`** - Original problem description and requirements

## Testing Strategy

The test suite covers:
1. **Node Creation** - Constructor function and structure
2. **Size Calculation** - Files, empty dirs, nested dirs
3. **Filesystem Building** - Command parsing and tree construction
4. **Directory Finding** - Size limits and filters
5. **Integration** - Full example with expected results

## Performance

- **Time Complexity**: O(n + d) where n = commands, d = directories
- **Space Complexity**: O(n) for tree structure

## Requirements

- Python 3.7+ (for dataclasses and type hints)
- pytest (for running tests)

```bash
# Install pytest if needed
pip install pytest
```

## Contributing

When modifying the solution:
1. Keep structural parity with Go implementation
2. Update tests to cover new functionality
3. Run tests to ensure nothing breaks: `python3 -m pytest test_solution.py -v`
4. Update documentation accordingly

## Comparison with Go

Both implementations produce **identical output** and use the **same algorithms**:

```bash
# Go
$ cd interview && go run . --inputFile input.txt
[{user 31673} {home 14880187} {log 584} {var 584} {/ 14880771}]
32841

# Python
$ cd interview && python3 solution.py --inputFile input.txt
[DirInfo(path='user', size=31673), DirInfo(path='home', size=14880187), DirInfo(path='log', size=584), DirInfo(path='var', size=584), DirInfo(path='/', size=14880771)]
32841
```

See `GO_PYTHON_COMPARISON.md` for a detailed side-by-side comparison.
