package main

import (
	"encoding/json"
	"fmt"
)

//func main() {
//	//1 i,v是两个变量，每次遍历数组对其赋值，所以地址不变
//	data:=[3]string{"a","b","c"}
//	for i,v:=range data{
//		println(&i,&v)
//	}
//	//2 数组的地址即是数组第一个元素的地址
//	fmt.Printf("%p\n",&data)
//	fmt.Printf("%p\n",&data[0])
//	fmt.Printf("%p\n",&data[1])
//
//}
//----------------------------------------------next--------------
//type Stu struct {
//	Name  string `json:"name"`
//	Age   int
//	HIgh  bool
//	sex   string
//	Class *Class `json:"class"`
//}
//
//type Class struct {
//	Name  string
//	Grade int
//}
//
//func main() {
//	//实例化一个数据结构，用于生成json字符串
//	stu := Stu{
//		Name: "张三",
//		Age:  18,
//		HIgh: true,
//		sex:  "男",
//	}
//
//	//指针变量
//	cla := new(Class)
//	cla.Name = "1班"
//	cla.Grade = 3
//	stu.Class=cla
//
//	//Marshal失败时err!=nil
//	jsonStu, err := json.Marshal(stu)
//	if err != nil {
//		fmt.Println("生成json字符串错误")
//	}
//
//	//jsonStu是[]byte类型，转化成string类型便于查看
//	fmt.Println(string(jsonStu))
//	fmt.Println(jsonStu)
//}
//-------------------------------next--------------------
type Employment struct {
	Company  string `json:"company"`//注释 -json后为company
	Title    string `json:"title"`
}
type Foo struct {
	Name  string                 `json:"name"`
	Age   int                    `json:"age"`
	Job   Employment             `json:"job"`//结构体
	Extra map[string]interface{} `json:"extra"`//接收未知数据类型
}

func main() {
	data := []byte(`{
        "name": "John Doe",
        "age": 25,
        "job": {
            "company": "ABC",
            "title": "Engineer"
        },
        "extra": {
            "marital status": "married",
            "childrens": 0
        }}`)
	var f Foo

	json.Unmarshal(data, &f)//把data数据解析到f里面
	fmt.Printf("%s is %d years old and works at %s.\n", f.Name, f.Age, f.Job.Company)
	fmt.Printf("He is %s and has %d childrens.\n", f.Extra["marital status"].(string), int(f.Extra["childrens"].(float64)))
	/*例子里的childrens字段值虽然是整数，但是被解析成了float64类型。在Json的定义中，值的类型不分整型和浮点型，
	只有一个number类型，因此在我们没有指定字段和类型的情况下，Unmarshal()函数把所有number类型的值都当作float64*/
//func Marshal(v interface{}) ([]byte, error) {
	output, _ := json.Marshal(&f)
	fmt.Println(string(output))
}