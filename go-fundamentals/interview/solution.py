#!/usr/bin/env python3
"""
Filesystem Directory Size Calculator

This script parses filesystem commands and calculates the total size
of each directory including all nested files and subdirectories.

Structure mirrors the Go implementation with standalone functions.
"""

import argparse
import os
from dataclasses import dataclass
from typing import Dict, List, Optional


class Node:
    """Represents a file or directory in the filesystem."""
    
    def __init__(self, name: str, is_dir: bool = False, size: int = 0):
        self.name = name
        self.is_dir = is_dir
        self.size = size
        self.children: Dict[str, 'Node'] = {}
        self.parent: Optional['Node'] = None


def new_node(name: str, is_dir: bool, size: int) -> Node:
    """
    Create a new Node instance.
    Mirrors Go's NewNode constructor.
    """
    return Node(name, is_dir, size)


def build_file_system(input_text: str) -> Node:
    """
    Build filesystem tree from input commands.
    Mirrors Go's buildFileSystem function.
    
    Args:
        input_text: Raw input string containing filesystem commands
        
    Returns:
        Root node of the filesystem tree
    """
    lines = input_text.strip().split('\n')
    
    root = new_node("/", True, 0)
    current = root
    
    for line in lines:
        parts = line.split()
        
        if not parts:
            continue
        
        # All commands contain only two strings
        if parts[0] == "cd":
            dirname = parts[1]
            if dirname == "/":
                current = root
            elif dirname == "..":
                if current.parent is not None:
                    current = current.parent
            else:
                if dirname in current.children:
                    # Switch parent to the new dir
                    current = current.children[dirname]
                else:
                    print(f"Directory: {dirname} doesn't exist")
        
        elif parts[0] == "dir":
            dirname = parts[1]
            if dirname not in current.children:
                new_dir = new_node(dirname, True, 0)
                new_dir.parent = current
                current.children[dirname] = new_dir
        
        else:
            # File <size> <filename>
            try:
                size = int(parts[0])
                filename = parts[1]
                
                file_node = new_node(filename, False, size)
                file_node.parent = current
                current.children[filename] = file_node
            except (ValueError, IndexError):
                print("Invalid file size")
    
    return root


def calculate_size(node: Node) -> int:
    """
    Recursively calculate the size of a node.
    For files, returns the file size.
    For directories, returns the sum of all children.
    Mirrors Go's calculateSize function.
    
    Args:
        node: The node to calculate size for
        
    Returns:
        Total size in bytes
    """
    if not node.is_dir:
        return node.size
    
    total = 0
    for child in node.children.values():
        total += calculate_size(child)
    
    return total


def find_dirs_under_limit(node: Node, limit: int) -> List[int]:
    """
    Find all directories with total size under a specified limit.
    Mirrors Go's findDirsUnderLimit function.
    
    Args:
        node: Starting node (usually root)
        limit: Size limit threshold
        
    Returns:
        List of directory sizes that are under the limit
    """
    if not node.is_dir:
        return []
    
    result = []
    current_size = 0
    
    for child in node.children.values():
        if child.is_dir:
            result.extend(find_dirs_under_limit(child, limit))
            current_size += calculate_size(child)
        else:
            current_size += child.size
    
    if current_size <= limit:
        result.append(current_size)
    
    return result


@dataclass
class DirInfo:
    """
    Directory information containing path and size.
    Mirrors Go's DirInfo struct.
    """
    path: str
    size: int
    
    def __repr__(self) -> str:
        return f"DirInfo(path='{self.path}', size={self.size})"


def find_dirs_size(node: Node, name: str) -> List[DirInfo]:
    """
    Return the directory names and the size for each of them.
    Mirrors Go's findDirsSize function.
    
    Args:
        node: Current node to process
        name: Name/path of the current node
        
    Returns:
        List of DirInfo objects containing directory information
    """
    if not node.is_dir:
        return []
    
    result = []
    current_size = 0
    
    # Calculate current directory size and collect subdirectory info
    for child in node.children.values():
        if child.is_dir:
            # Recursively collect info from subdirectories
            result.extend(find_dirs_size(child, child.name))
            # Add subdirectory size to current
            current_size += calculate_size(child)
        else:
            # Add file size to current directory
            current_size += child.size
    
    # Add current directory info
    result.append(DirInfo(path=name, size=current_size))
    
    return result


def run(input_file: str = "input.txt") -> None:
    """
    Main execution function.
    Mirrors Go's Run function.
    
    Args:
        input_file: Path to input file
    """
    try:
        with open(input_file, 'r') as f:
            input_text = f.read()
    except FileNotFoundError:
        print(f"Unable to read the input file: {input_file}")
        return
    
    root = build_file_system(input_text)
    
    dirs = find_dirs_size(root, "/")
    print(dirs)
    
    sizes = find_dirs_under_limit(root, 100000)
    
    total_sum = sum(sizes)
    print(total_sum)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Filesystem Directory Size Calculator")
    parser.add_argument(
        "--inputFile",
        type=str,
        default="input.txt",
        help="Relative path to the input file (default: input.txt)"
    )
    
    args = parser.parse_args()
    
    # If input file is not absolute, make it relative to script directory
    if not os.path.isabs(args.inputFile):
        script_dir = os.path.dirname(os.path.abspath(__file__))
        input_file = os.path.join(script_dir, args.inputFile)
    else:
        input_file = args.inputFile
    
    run(input_file)
