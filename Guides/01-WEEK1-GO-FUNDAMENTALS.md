# Week 1: Go Fundamentals - 40 Problems

## Goal
Master Go syntax, idioms, and standard library through 40 carefully selected problems:
- **20 LeetCode problems** (covering data structures & algorithms)
- **20 Advent of Code problems** (real-world problem solving)

---

## LeetCode Problems (20)

### Day 1-2: Arrays, Slices, Maps, Strings (7 problems)

**Go Focus**: Slices vs arrays, maps, range loops, rune handling

1. **Two Sum** - [LeetCode #1](https://leetcode.com/problems/two-sum/)
2. **Valid Anagram** - [LeetCode #242](https://leetcode.com/problems/valid-anagram/)
3. **Contains Duplicate** - [LeetCode #217](https://leetcode.com/problems/contains-duplicate/)
4. **Product of Array Except Self** - [LeetCode #238](https://leetcode.com/problems/product-of-array-except-self/)
5. **Group Anagrams** - [LeetCode #49](https://leetcode.com/problems/group-anagrams/)
6. **Longest Substring Without Repeating Characters** - [LeetCode #3](https://leetcode.com/problems/longest-substring-without)
7. **Valid Palindrome** - [LeetCode #125](https://leetcode.com/problems/valid-palindrome/)

### Day 3: Structs, Methods, Interfaces (4 problems)

**Go Focus**: Pointer receivers, interfaces, custom types

8. **LRU Cache** - [LeetCode #146](https://leetcode.com/problems/lru-cache/)
9. **Min Stack** - [LeetCode #155](https://leetcode.com/problems/min-stack/)
   - Difficulty: Medium
   - Concepts: Stack interface, error handling

10. **Implement Queue using Stacks** - [LeetCode #232](https://leetcode.com/problems/implement-queue-using-stacks/)
    - Difficulty: Easy
    - Concepts: Interface implementation, methods

11. **Design HashSet** - [LeetCode #705](https://leetcode.com/problems/design-hashset/)
    - Difficulty: Easy
    - Concepts: Custom data structures, methods

### Day 4-5: Sorting, Searching, Binary Search (4 problems)

**Go Focus**: sort package, sort.Interface, binary search

12. **Binary Search** - [LeetCode #704](https://leetcode.com/problems/binary-search/)
    - Difficulty: Easy
    - Concepts: Binary search algorithm

13. **Merge Intervals** - [LeetCode #56](https://leetcode.com/problems/merge-intervals/)
    - Difficulty: Medium
    - Concepts: Custom sorting with sort.Slice

14. **Kth Largest Element in an Array** - [LeetCode #215](https://leetcode.com/problems/kth-largest-element-in-an-array/)
    - Difficulty: Medium
    - Concepts: container/heap package, heaps

15. **Search in Rotated Sorted Array** - [LeetCode #33](https://leetcode.com/problems/search-in-rotated-sorted-array/)
    - Difficulty: Medium
    - Concepts: Modified binary search

### Day 6: Recursion, Backtracking (3 problems)

**Go Focus**: Slice copying, recursive patterns, backtracking

16. **Permutations** - [LeetCode #46](https://leetcode.com/problems/permutations/)
    - Difficulty: Medium
    - Concepts: Backtracking, slice manipulation

17. **Combination Sum** - [LeetCode #39](https://leetcode.com/problems/combination-sum/)
    - Difficulty: Medium
    - Concepts: Recursion, slice operations

18. **Generate Parentheses** - [LeetCode #22](https://leetcode.com/problems/generate-parentheses/)
    - Difficulty: Medium
    - Concepts: String building, recursion

### Day 7: Stack, Queue, Linked List (2 problems)

**Go Focus**: Pointers, nil handling, data structure implementation

19. **Valid Parentheses** - [LeetCode #20](https://leetcode.com/problems/valid-parentheses/)
    - Difficulty: Easy
    - Concepts: Stack pattern, error handling

20. **Reverse Linked List** - [LeetCode #206](https://leetcode.com/problems/reverse-linked-list/)
    - Difficulty: Easy
    - Concepts: Pointers, linked lists, nil handling

---

## Advent of Code Problems (20)
### Day 1-2: String Parsing, File I/O (4 problems)

21. **AoC 2020 Day 1** - https://adventofcode.com/2020/day/1
22. **AoC 2022 Day 1** - https://adventofcode.com/2022/day/1
23. **AoC 2021 Day 1** - https://adventofcode.com/2021/day/1
24. **AoC 2023 Day 1** - https://adventofcode.com/2023/day/1

### Day 3: Maps, Counting, Frequency (2 problems)

25. **AoC 2021 Day 3** - [Binary Diagnostic](https://adventofcode.com/2021/day/3)
26. **AoC 2022 Day 6** - [Tuning Trouble](https://adventofcode.com/2022/day/6)

### Day 4: Grid Problems, 2D Slices (3 problems)
**Go Focus**: 2D slices, coordinate systems
27. **AoC 2021 Day 4** - [Giant Squid](https://adventofcode.com/2021/day/4)
28. **AoC 2022 Day 8** - [Treetop Tree House](https://adventofcode.com/2022/day/8) -> TODO
    - Grid visibility checking
    - Multiple directions
    - Concepts: 2D iteration, directions
29. **AoC 2023 Day 3** - [Gear Ratios](https://adventofcode.com/2023/day/3)

### Day 5: Structs, Custom Types (3 problems)

**Go Focus**: Struct design, methods, type definitions

30. **AoC 2021 Day 5** - [Hydrothermal Venture](https://adventofcode.com/2021/day/5)
31. **AoC 2022 Day 4** - [Camp Cleanup](https://adventofcode.com/2022/day/4)
32. **AoC 2020 Day 4** - [Passport Processing](https://adventofcode.com/2020/day/4)

## Day 6: System Design problems (4 problems)
**AoC 2022 Day 11** - [Monkey in the Middle](https://adventofcode.com/2022/day/11)
**AoC 2022 Day 7** - [File System](https://adventofcode.com/2022/day/7)

### Day 7: Algorithms, Recursion (3 problems)

**Go Focus**: Memoization, recursion, dynamic programming

33. **AoC 2021 Day 6** - [Lanternfish](https://adventofcode.com/2021/day/6)
    - Population simulation
    - Efficient counting
    - Concepts: Maps for memoization

34. **AoC 2020 Day 7** - [Handy Haversacks](https://adventofcode.com/2020/day/7)
    - Graph traversal
    - Bag containment rules
    - Concepts: Recursion, graph representation


### Day 7: Advanced Parsing, State Machines (5 problems)

**Go Focus**: JSON, complex parsing, state management

36. **AoC 2020 Day 2** - [Password Philosophy](https://adventofcode.com/2020/day/2)
    - Parse structured rules
    - Validation logic
    - Concepts: regexp, strings.Fields

37. **AoC 2021 Day 2** - [Dive!](https://adventofcode.com/2021/day/2)
    - Command parsing
    - State updates
    - Concepts: State machine, switch statements

38. **AoC 2022 Day 5** - [Supply Stacks](https://adventofcode.com/2022/day/5)
    - Stack operations
    - Complex input parsing
    - Concepts: Stack data structure, parsing

39. **AoC 2023 Day 2** - [Cube Conundrum](https://adventofcode.com/2023/day/2)
    - Game state parsing
    - Min/max tracking
    - Concepts: Parsing, aggregation

40. **AoC 2021 Day 7** - [The Treachery of Whales](https://adventofcode.com/2021/day/7)
    - Optimization problem
    - Median/mean calculation
    - Concepts: Sorting, math operations

## Project Structure

```
01-go-fundamentals/
├── go.mod
├── leetcode/
│   ├── 001_two_sum/
│   │   ├── solution.go
│   │   └── solution_test.go
│   ├── 242_valid_anagram/
│   │   ├── solution.go
│   │   └── solution_test.go
│   └── ...
├── aoc/
│   ├── 2020_day01/
│   │   ├── input.txt
│   │   ├── solution.go
│   │   └── solution_test.go
│   ├── 2022_day01/
│   │   ├── input.txt
│   │   ├── solution.go
│   │   └── solution_test.go
│   └── ...
└── utils/
    ├── input.go      // Common input helpers
    └── math.go       // Math utilities
```
