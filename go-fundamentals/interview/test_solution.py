#!/usr/bin/env python3
"""
Test suite for the filesystem directory size calculator.
Mirrors the Go implementation structure.
"""

import unittest
from solution import (
    Node,
    new_node,
    build_file_system,
    calculate_size,
    find_dirs_under_limit,
    find_dirs_size,
    DirInfo
)


class TestNode(unittest.TestCase):
    """Test cases for the Node class and new_node function."""
    
    def test_new_node_file(self):
        """Test creating a file node."""
        node = new_node("test.txt", False, 100)
        self.assertEqual(node.name, "test.txt")
        self.assertFalse(node.is_dir)
        self.assertEqual(node.size, 100)
    
    def test_new_node_directory(self):
        """Test creating a directory node."""
        node = new_node("testdir", True, 0)
        self.assertEqual(node.name, "testdir")
        self.assertTrue(node.is_dir)
        self.assertEqual(len(node.children), 0)
    
    def test_node_parent_child_relationship(self):
        """Test parent-child relationships."""
        parent = new_node("parent", True, 0)
        child = new_node("child.txt", False, 50)
        
        child.parent = parent
        parent.children["child.txt"] = child
        
        self.assertIn("child.txt", parent.children)
        self.assertEqual(child.parent, parent)


class TestCalculateSize(unittest.TestCase):
    """Test cases for the calculate_size function."""
    
    def test_calculate_size_file(self):
        """Test calculating size of a file node."""
        file_node = new_node("file.txt", False, 100)
        self.assertEqual(calculate_size(file_node), 100)
    
    def test_calculate_size_empty_directory(self):
        """Test calculating size of an empty directory."""
        dir_node = new_node("dir", True, 0)
        self.assertEqual(calculate_size(dir_node), 0)
    
    def test_calculate_size_directory_with_files(self):
        """Test calculating size of a directory with files."""
        dir_node = new_node("dir", True, 0)
        file1 = new_node("file1.txt", False, 100)
        file2 = new_node("file2.txt", False, 200)
        
        file1.parent = dir_node
        file2.parent = dir_node
        dir_node.children["file1.txt"] = file1
        dir_node.children["file2.txt"] = file2
        
        self.assertEqual(calculate_size(dir_node), 300)
    
    def test_calculate_size_nested_directories(self):
        """Test calculating size of nested directories."""
        root = new_node("root", True, 0)
        subdir = new_node("subdir", True, 0)
        file1 = new_node("file1.txt", False, 100)
        file2 = new_node("file2.txt", False, 50)
        
        file1.parent = root
        subdir.parent = root
        file2.parent = subdir
        
        root.children["file1.txt"] = file1
        root.children["subdir"] = subdir
        subdir.children["file2.txt"] = file2
        
        self.assertEqual(calculate_size(root), 150)
        self.assertEqual(calculate_size(subdir), 50)


class TestBuildFileSystem(unittest.TestCase):
    """Test cases for the build_file_system function."""
    
    def test_build_simple_filesystem(self):
        """Test building a simple filesystem."""
        input_text = """cd /
dir home
100 file1.txt"""
        
        root = build_file_system(input_text)
        
        self.assertEqual(root.name, "/")
        self.assertIn("home", root.children)
        self.assertIn("file1.txt", root.children)
        self.assertEqual(root.children["file1.txt"].size, 100)
    
    def test_build_filesystem_with_navigation(self):
        """Test building filesystem with directory navigation."""
        input_text = """cd /
dir home
cd home
200 file2.txt"""
        
        root = build_file_system(input_text)
        
        self.assertIn("home", root.children)
        home_dir = root.children["home"]
        self.assertIn("file2.txt", home_dir.children)
        self.assertEqual(home_dir.children["file2.txt"].size, 200)
    
    def test_build_filesystem_with_parent_navigation(self):
        """Test building filesystem with parent directory navigation."""
        input_text = """cd /
dir home
cd home
100 file1.txt
cd ..
200 file2.txt"""
        
        root = build_file_system(input_text)
        
        self.assertIn("file2.txt", root.children)
        self.assertIn("home", root.children)
        home_dir = root.children["home"]
        self.assertIn("file1.txt", home_dir.children)


class TestFindDirsUnderLimit(unittest.TestCase):
    """Test cases for the find_dirs_under_limit function."""
    
    def test_find_dirs_under_limit_simple(self):
        """Test finding directories under a limit."""
        root = new_node("/", True, 0)
        dir1 = new_node("dir1", True, 0)
        file1 = new_node("file1.txt", False, 50)
        
        dir1.parent = root
        file1.parent = dir1
        root.children["dir1"] = dir1
        dir1.children["file1.txt"] = file1
        
        result = find_dirs_under_limit(root, 100)
        
        self.assertIn(50, result)
    
    def test_find_dirs_under_limit_exceeds(self):
        """Test that directories over limit are not included."""
        root = new_node("/", True, 0)
        dir1 = new_node("dir1", True, 0)
        file1 = new_node("file1.txt", False, 150)
        
        dir1.parent = root
        file1.parent = dir1
        root.children["dir1"] = dir1
        dir1.children["file1.txt"] = file1
        
        result = find_dirs_under_limit(root, 100)
        
        self.assertNotIn(150, result)


class TestFindDirsSize(unittest.TestCase):
    """Test cases for the find_dirs_size function."""
    
    def test_find_dirs_size_simple(self):
        """Test finding directory sizes."""
        root = new_node("/", True, 0)
        file1 = new_node("file1.txt", False, 100)
        
        file1.parent = root
        root.children["file1.txt"] = file1
        
        result = find_dirs_size(root, "/")
        
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].path, "/")
        self.assertEqual(result[0].size, 100)
    
    def test_find_dirs_size_nested(self):
        """Test finding sizes of nested directories."""
        root = new_node("/", True, 0)
        dir1 = new_node("dir1", True, 0)
        file1 = new_node("file1.txt", False, 100)
        file2 = new_node("file2.txt", False, 50)
        
        dir1.parent = root
        file1.parent = root
        file2.parent = dir1
        
        root.children["dir1"] = dir1
        root.children["file1.txt"] = file1
        dir1.children["file2.txt"] = file2
        
        result = find_dirs_size(root, "/")
        
        self.assertEqual(len(result), 2)
        
        # Find dir1 and root in results
        dir1_info = next(d for d in result if d.path == "dir1")
        root_info = next(d for d in result if d.path == "/")
        
        self.assertEqual(dir1_info.size, 50)
        self.assertEqual(root_info.size, 150)


class TestExampleInput(unittest.TestCase):
    """Test with the full example from the problem."""
    
    def test_example_input(self):
        """Test with the example input from the problem."""
        input_text = """cd /
dir home
dir var
cd home
dir user
14848514 largefile.txt
cd user
29116 photo.jpg
2557 config.txt
cd ..
cd ..
cd var
dir log
cd log
584 system.log"""
        
        root = build_file_system(input_text)
        dirs = find_dirs_size(root, "/")
        
        # Convert to dict for easier testing
        dir_dict = {d.path: d.size for d in dirs}
        
        # Verify expected sizes
        self.assertEqual(dir_dict["user"], 31673)
        self.assertEqual(dir_dict["home"], 14880187)
        self.assertEqual(dir_dict["log"], 584)
        self.assertEqual(dir_dict["var"], 584)
        self.assertEqual(dir_dict["/"], 14880771)
    
    def test_find_dirs_under_limit_example(self):
        """Test finding directories under 100000 with example input."""
        input_text = """cd /
dir home
dir var
cd home
dir user
14848514 largefile.txt
cd user
29116 photo.jpg
2557 config.txt
cd ..
cd ..
cd var
dir log
cd log
584 system.log"""
        
        root = build_file_system(input_text)
        sizes = find_dirs_under_limit(root, 100000)
        
        # Should find user (31673), log (584), and var (584)
        self.assertIn(31673, sizes)
        self.assertIn(584, sizes)
        
        # Total sum should match
        total = sum(sizes)
        # user: 31673, log: 584, var: 584 = 32841
        self.assertEqual(total, 32841)


if __name__ == "__main__":
    unittest.main()
