package main

import "io/ioutil"
import "encoding/json"

type Settings struct{
    Name string
    Password string
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