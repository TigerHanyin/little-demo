package main

import "fmt"

func main() {
	//1 i,v是两个变量，每次遍历数组对其赋值，所以地址不变
	data:=[3]string{"a","b","c"}
	for i,v:=range data{
		println(&i,&v)
	}
	//2 数组的地址即是数组第一个元素的地址
	fmt.Printf("%p\n",&data)
	fmt.Printf("%p\n",&data[0])
	fmt.Printf("%p\n",&data[1])

}
