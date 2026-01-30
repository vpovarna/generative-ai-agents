## Go Concepts

**struct{}** takes 0 bytes of memory. Where if I would use **bool**, that would take 1 byte of memory.  

To create a set: map[Type]bool  

To check if an element is a set: _, exist := mySet[key]. Checking an element and returning the value, and doing something with the returned values can be one line: 
    if j, exist := seen[component]; exist {
    }  

To print arrays: fmt.Printf("%v\n", nums)

To test a function:
func TestFunctionName(t *testing.T) {
    // Test code here
}

Key Testing Functions: 
- t.Error() / t.Errorf() - Report error but continue test
- t.Fatal() / t.Fatalf() - Report error and stop test immediately
- t.Log() / t.Logf() - Log information(only shown with -v flag)
- t.Run() - run subtests

A test function must start with Test and should take *testing.T parameter. 
't' is a pointer to testing.T . It gives you methods to control the test  


