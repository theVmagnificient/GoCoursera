package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"syscall"
)

func elInSlice(a int, list []int) (bool, int) {
	for idx, b := range list {
		if b == a {
			return true, idx
		}
	}
	return false, -1
}

type OsInstance struct {
	Depth int
	Path string
	Children []*OsInstance
}

func (o *OsInstance) DirTree(vert []int, symb string, isLast bool) {

	str := ""
	i := 0
	for i < o.Depth {
		if flag, _ := elInSlice(i, vert); flag {
			if i == vert[len(vert) - 1] {
				str += symb
			} else {
				str += "|"
			}
		} else if i > o.Depth - 4 {
			str += "─"
		} else {
			str += " "
		}
		i++
	}

	parts := strings.Split(o.Path, "/")

	fmt.Printf(str + parts[len(parts) - 1] + "\n")



	idx := 0
	vert = append(vert, vert[len(vert) - 1] + 4)

	new_vert := make([]int, len(vert)-1, 10) // len=5, cap=10
	if isLast {
		flag, ix := elInSlice(o.Depth - 4, vert)

		cur_len := 0
		if flag == true {
			for index, val := range vert {
				if index != ix {
					new_vert[cur_len] = val
					cur_len++
				}
			}
		}
	} else {
		new_vert = vert
	}

	for idx < len(o.Children) {
		if idx != len(o.Children) - 1 {
			o.Children[idx].DirTree(new_vert, "├", false)
		} else {
			o.Children[idx].DirTree(new_vert, "└", true)
		}
		idx++
	}

}


func (o *OsInstance) GetSortedChildren() error {
	children, err := ioutil.ReadDir(o.Path)
	if err != nil {
		file, err := os.Open(o.Path)

		fi, err := file.Stat();

		if !fi.IsDir() {
			return nil
		}

		errno := uintptr(err.(*os.SyscallError).Err.(syscall.Errno))

		if errno != 20 { // Error when trying to read file as dir
			fmt.Println("Shit happened in %v", o.Path)
			return err
		} else {
			return nil
		}
	}
	for _, f := range children {
		o.Children = append(o.Children,&(OsInstance{Path:path.Join(o.Path, f.Name()),
			                                        Depth: o.Depth + 4, Children:nil}))
	}

	if children != nil {
		sort.Slice(o.Children, func(i, j int) bool {
			return o.Children[i].Path < o.Children[j].Path
		})

		id := 0
		for id < len(o.Children) {
			err = o.Children[id].GetSortedChildren()
			if err != nil {
				fmt.Println(err)
				return err
			}
			id++
		}
	}
	return nil
}

func main() {

	var start OsInstance = OsInstance{Depth: 4, Path:"/home/v/Dev/Go"}

	if err := start.GetSortedChildren(); err != nil {
		fmt.Println(err)
	}
	start.DirTree([]int{0}, "├", true)

}
