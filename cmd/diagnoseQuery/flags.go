package main

import (
	"diagnoseQuery/internal/parameter"
	"fmt"
)

func PrintHelp() {
	fmt.Println("필수 옵션을 입력하지 않았습니다.")
	fmt.Println("")
	fmt.Println("사용법:")
	fmt.Println("    diagnoseQuery --query '<SQL문>': 지정한 SQL문 실행")
	fmt.Println("    diagnoseQuery --explain '<SQL문>': 쿼리 성능 분석")
	fmt.Println("\n예시:")
	fmt.Println("    diagnoseQuery --query 'SELECT * FROM resources WHERE server_id = 1'")
	fmt.Println("    diagnoseQuery --explain 'SELECT * FROM resources WHERE server_id = 1'")
	fmt.Println("\n주의: --query 또는 --explain 옵션을 사용할 때는 SQL문을 따옴표로 감싸야 합니다.")
	fmt.Println("주의: --explain 옵션은 SELECT 쿼리만 지원합니다.")
}

func handleEnvPathOption(args []string, store *parameter.InMemoryStore) error {
	if len(args) < 1 {
		return fmt.Errorf("--env-path 사용 시 경로를 인자로 입력하세요.")
	}
	envPath := args[0]
	return parameter.LoadEnvToMemory(envPath, store)
}

func handleEnvsOption(args []string, store *parameter.InMemoryStore) error {
	if len(args) < 1 {
		return fmt.Errorf("--envs 사용 시 키를 인자로 입력하세요.")
	}
	key := args[0]
	return parameter.LoadKeyToMemory(key, store)
}

func handleDefaultOption(args []string, store *parameter.InMemoryStore) error {
	if len(args) < 1 {
		return fmt.Errorf("기본 옵션 사용 시 키를 인자로 입력하세요.")
	}
	key := args[0]
	return parameter.InputEnvToMemory(key, "default_value", store)
}

func processOption(option string, args []string, store *parameter.InMemoryStore) error {
	switch option {
	case "--env-path":
		return handleEnvPathOption(args, store)
	case "--envs":
		return handleEnvsOption(args, store)
	default:
		return handleDefaultOption(args, store)
	}
}

func ProcessFlags(args []string, store *parameter.InMemoryStore) error {
	for i := 0; i < len(args); i++ {
		option := args[i]
		var nextArgs []string
		if i+1 < len(args) {
			nextArgs = args[i+1:]
		}
		if err := processOption(option, nextArgs, store); err != nil {
			return fmt.Errorf("옵션 처리 실패: %v", err)
		}
		if option == "--env-path" || option == "--envs" {
			i++
		}
	}
	return nil
}
