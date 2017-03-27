package main

import "fmt"

func main() {
    settings:=get_settings("settings.json")
    tclient:=NewTravianClient()
    tdata:=TravianData{}
    dorf1:=tclient.login(settings.Name, settings.Password)
    tdata.gather_data_from_dorf1(dorf1)
    fmt.Println(tdata)
}