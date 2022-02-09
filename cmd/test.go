package main

import (
    "fmt" 
    "math/rand"
    "math"
    "runtime"
    "time"
    "strings"
)

var aVar, another bool = true, false
// k := 5               // := not allowed outside func

var (
    multiple = 5
    vars = true
    inablock = "hello"
)

const Pi = 3.14         // starts with capital means it's exported 



func main() {
	fmt.Println("My favorite number is", rand.Intn(10))
	
	fmt.Printf("Now you have %g problems.\n", math.Sqrt(7))
	
	fmt.Println(math.Pi)
	
	fmt.Println(add(10, 5))
	
	fmt.Println(swap("world", "hello"))
	
	fmt.Println(split(20))
	
	fmt.Println(aVar, another)
	
	k := 5      // create and assign in one step with := inside a func
    var z int   // declare but not assign var
    z = 5
	
	fmt.Println("convert ", k, float64(k), uint(k))
	fmt.Println(z)
	
	sum := 0
	for i := 0; i < 10; i++ {               // NO () ALLOWED IN FOR OR IF
        sum += i
    }
    fmt.Println("sum is ", sum)
    
    if v:= 5; v<100 {             // declare a var before the condition, like a using clause
        fmt.Println("5 is < 100")
    } else {                            // ELSE MUST BE ON SAME LINE AS }
        // v is accessible here
        fmt.Println("5 >= 100???")
    }
    
	fmt.Println("sqrt ", Sqrt(4))
	
    basicSwitch()
	
	deferredFunction()
	
	pointers()
	
	structs()
	
	arrays()
	
	ranges()
	
	maps()
	
	functions()
	
	closures()
}

func add(x int, y int) int {
	return x + y
}

// return multiple values
func swap(x, y string) (string, string) {
	return y, x
}

// naked returns can assign values to named return values
// this sucks, don't do this 
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return
}

func Sqrt(x float64) float64 {
    z := float64(1)
    for i:=0; i<10; i++ {
        z -= (z*z - x) / (2*z)
    }
    return z
}

func basicSwitch() {
    // declare a variable os and then switch on it
    // cases don't fall through 
    fmt.Print("my os is ")
    switch os := runtime.GOOS; os {
        case "darwin":
            fmt.Println("mac")
        case "linux":
            fmt.Println("linux")
        default:
            fmt.Println("other", os)
    }
    
    today := time.Now().Weekday()
    fmt.Println("when's saturday?", today)
    switch time.Saturday {
        case today + 0:
            fmt.Println("Today")
        case today + 1:
            fmt.Println("Tomorrow")
        case today + 2:
            fmt.Println("In two days.")
        default:
            fmt.Println("too far")
    }
    
}

func deferredFunction() {
    defer fmt.Println("this is last")
    defer fmt.Print("world ")
    fmt.Print("deferred hello ")
}

func pointers() {
    var p *int
    i := 42
    p = &i      // get address of i
    *p = 21     // set the value by dereferencing
    fmt.Println("p is", *p, "i is", i)     // dereference p to get value
    
}

type Vertex struct {
    X int           // capital is exported
    Y int
}

func structs() {
    v := Vertex{1, 2}
    fmt.Println(v)   
    v.X = 4
    fmt.Println(v)
    p := &v
    p.X = 100           // can access struct fields through a pointer without dereferencing
    fmt.Println(*p)
    
    v2 := Vertex{X:1}       // Y is 0
    fmt.Println(v2)
    p = &Vertex{5, 6}       // & to create pointer to struct
}

func arrays() {
    var a [10]int           // arrays have fixed size, uninited is all 0
    fmt.Println(a)
    primes := [6]int{2, 3, 5, 7, 11, 13}      // initialize
    fmt.Println(primes)
    
    // slices have dynamic size, are backed by an array
    var s []int = primes[2:4]     // take a slice of primes with elements 2 (incl) through 4 (excl)
    fmt.Println(s)
    // changing slice value actually changes underlying array value

    // creating a slice literal also creates an array backing it
    var b = []bool{true, false, true}
    fmt.Println(b)    
    
    // can use a[:] to get slice of entire array a
    
    // len() of a slice is the number of elements in it
    // cap() of a slice is the number of elements from first slice element to last array element
    // can extend length if there is capacity
    var q = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} 
    q = q[:8]      // shorten to 8
    q = q[2:]
    fmt.Println(q, len(q), cap(q))
    
    // uninited slice has value nil, len and cap == 0
    
    // make(type, len, cap=0) creates uninited slice
    m := make([]int, 5)
    fmt.Println(m)
    
    // multidim slices
    board := [][]string{
        []string{"_", "_", "_"},
        []string{"_", "_", "_"},
        []string{"_", "_", "_"},
    }
    board[0][1] = "X"
    board[1][2] = "O"
    for i:=0; i<len(board); i++ {
        fmt.Println(strings.Join(board[i], " "))
    }
    
    // append(slice, vals...) returns new slice with appended values, possibly reallocating backing array to be larger
    var x = []int{1}
    x = append(x, 2, 3, 4)
    fmt.Println(x)
}

func ranges() {
    pow := []int{1, 2, 4, 8, 16, 32, 64, 128}
	for index, val := range pow {
		fmt.Printf("2**%d = %d\n", index, val)
	}
	// can ignore either index or val by assigning to _: 
	// for i, _ := range pow
	// can just get index: 
	// for i := range pow
}

func Pic(dx, dy int) [][]uint8 {
	r := make([][]uint8, dy)
	for i := range r {
		l := make([]uint8, dx)
		for j := range l {
			l[j] = uint8((i*i + j*j)/2)
		}
		r[i] = l
	}
	return r
}

var m map[string]Vertex         // creates an empty map of string -> Vertex. The zero value of a map is nil. A nil map has no keys, nor can keys be added.
func maps() {
    //The make function returns a map of the given type, initialized and ready for use.
    m = make(map[string]Vertex)
    m["BellLabs"] = Vertex{20, 30}
    fmt.Println(m)
    fmt.Println(m["BellLabs"])
    
    // map literal
    m = map[string]Vertex {
    	"Bell Labs": Vertex{ 40, -74, },                // can exclude the Vertex type here if wanted
    	"Google": Vertex{ 37, -122, },                      // TRAILING COMMAs ARE REQUIRED
    }
    fmt.Println(m)
    
    // delete a key 
    delete(m, "Bell Labs")
    
    // check for existence of key
    elem, ok := m["banana"]      // ok is bool
    fmt.Println("exists?", ok, elem)
}

func WordCount(s string) map[string]int {
	m := make(map[string]int)
	words := strings.Fields(s)
	for _, v := range words {
		m[v] = m[v] + 1
	}
	return m
}

func compute(fn func(float64, float64) float64) float64 {           // parameter is a function that takes 2 float64s and returns a float64. compute returns a float64
	return fn(3, 4)         // simply calls the given function with 3, 4 and returns the value
}

func functions() {
    // can declare functions as vars
    hypot := func(x, y float64) float64 {
        return math.Sqrt(x*x + y*y)
    }
    fmt.Println(hypot(5, 12))

    fmt.Println(compute(hypot))     // pass name of hypot var to compute
}

func adder() func(int) int {
    // this is a closure, each instance of this function will share the sum
    sum := 0
    return func(x int) int {
        sum += x
        return sum
    }
}

func closures() {
    pos, neg := adder(), adder()
    for i:=0; i<10; i++ {
        fmt.Println(pos(i), neg(-i))
    }
}
