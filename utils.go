package gonetplan

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type numSort struct {
	name  string
	index int
}

// 文件排序规则
type byNum []numSort

func (a byNum) Len() int           { return len(a) }
func (a byNum) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byNum) Less(i, j int) bool { return a[i].index < a[j].index }

func extractNum(filename string) (int, error) {
	parts := strings.Split(filename, "-")
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil {
			return num, nil
		}
	}
	return 0, fmt.Errorf("no numeric part found in %s", filename)
}

func OrderFiles(files []string) []string {
	numFiles := make([]numSort, len(files))
	for i, file := range files {
		num, err := extractNum(filepath.Base(file))
		if err != nil {
			fmt.Println(err)
			continue
		}
		numFiles[i] = numSort{name: file, index: num}
	}

	// 对文件进行排序
	sort.Sort(byNum(numFiles))
	var returnfile []string
	for _, file := range numFiles {
		returnfile = append(returnfile, file.name)
	}
	return returnfile
}

func GetMapKeys(m map[string]interface{}) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ApplyNetplanConfig() error {
	if err := exec.Command("netplan", "apply").Run(); err != nil {
		return err
	}
	return nil
}

func RemoveVirtualDevices(devName string) error {
	if err := exec.Command("ip", "link", "delete", devName).Run(); err != nil {
		return err
	}
	return nil
}

// difference 返回sliceB中存在，但sliceA中不存在的元素
func difference[T comparable](sliceA, sliceB []T) []T {
	// 创建一个映射，用于存储sliceA中的所有元素
	m := make(map[T]bool)
	// 将sliceA中的元素添加到映射中
	for _, item := range sliceA {
		m[item] = true
	}

	// 创建一个新的slice，用于存储仅存在于sliceB中的元素
	var diff []T
	for _, item := range sliceB {
		// 如果sliceA中不存在该元素，则添加到结果slice中
		if _, found := m[item]; !found {
			diff = append(diff, item)
		}
	}

	return diff
}
