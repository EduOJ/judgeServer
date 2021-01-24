package main

import (
	"fmt"
	"github.com/suntt2019/Judger"
)

func main() {
	config := judger.Config{
		MaxCPUTime:           1000,
		MaxRealTime:          2000,
		MaxMemory:            128 * 1024 * 1024,
		MaxStack:             32 * 1024 * 1024,
		MaxProcessNumber:     200,
		MaxOutputSize:        10000,
		MemoryLimitCheckOnly: 0,
		ExePath:              "test_programs/a+b/a+b",
		InputPath:            "test_programs/a+b/a+b.in",
		OutputPath:           "test_programs/a+b/a+b.out",
		ErrorPath:            "test_programs/a+b/a+b.out",
		Args:                 []string{},
		Env:                  []string{},
		LogPath:              "test.log",
		SeccompRuleName:      "c_cpp",
		Uid:                  0,
		Gid:                  0,
	}
	result := judger.Run(config)
	fmt.Printf("CPUTime: %v\n", result.CPUTime)
	fmt.Printf("RealTime: %v\n", result.RealTime)
	fmt.Printf("Memory: %v\n", result.Memory)
	fmt.Printf("Signal: %v\n", result.Signal)
	fmt.Printf("ExitCode: %v\n", result.ExitCode)
	fmt.Printf("Result: %v\n", result.Result)
	fmt.Printf("Error: %v\n", result.Error)
}
