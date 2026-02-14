### Problem
Problem Description
You are given a series of commands that represent operations on a file system. Your task is to:
- Find all directories and calculate the sum for each one of them. 

Input Format  
The input consists of lines with the following commands:  
- `cd <dirname>` - Navigate into a directory  
- `cd ..` - Navigate to parent directory  
- `dir <dirname>` - Declare a subdirectory exists    
- `<size> <filename>` - A file with the given size
Note: The root directory is /

Input:
```
cd /
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
584 system.log
```

Expected Output:
```
user -> 31673
home -> 14880187
log -> 584 
var -> 584 
/ -> 14880771
```

Directory tree:
```
/
├── home/
│   ├── user/
│   │   ├── photo.jpg (29116)
│   │   └── config.txt (2557)
│   └── largefile.txt (14848514)
└── var/
    └── log/
        └── system.log (584)
```

### Scoring
#### Junior Engineer

- Node struct
- buildFileSystem
- calculateSize() - working recursively


#### Mid-Level Engineer
- Node struct
- buildFileSystem
- calculateSize() - working recursively
- findDirsSize() - WORKING. Returns correct DirInfo slice

#### Senior-Level Engineer
- Working problem
- Adding Circular Reference case.
