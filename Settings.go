package main

import "io/ioutil"
import "encoding/json"

type Settings struct{
    Server string
    Name string
    Password string
    Fields_p1 []int
    Fields_p2 []int
    Fields_p3 []int
}

func get_settings(filename string) Settings {
    content, err:=ioutil.ReadFile(filename)
    if err!=nil{
        panic(err.Error())
    }
    settings:=Settings{}
    err=json.Unmarshal(content, &settings)
    if err!=nil{
        panic(err.Error())
    }
    return settings
}