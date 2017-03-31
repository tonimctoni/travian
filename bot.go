package main

import "log"
import "os"
import "time"
import "math/rand"
import "fmt"
// o877558
// o1004120
func main() {
    settings:=get_settings("settings.json")
    tclient:=NewTravianClient()
    tclient.login(settings.Url, settings.Name, settings.Password)

    file, err:=os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err!=nil{
        panic(err.Error())
    }
    defer file.Close()
    mylog:=log.New(file, "", log.LstdFlags)

    mylog.Println("Starting")
    outer: for{
        err:=tclient.gather_data_from_dorf1()
        if err!=nil{
            fmt.Println(err.Error())
        }
        if tclient.tdata.is_logged_in{
            if !tclient.tdata.is_upgrading{
                for _, i:=range rand.Perm(18){ //int64(1);i<=18;i++{
                    i+=1
                    fmt.Println(i)
                    able, err:=tclient.try_upgrade_farm(i)
                    if err!=nil{
                        mylog.Println("Farm upgrade failed for farm", i, "...", err.Error())
                        continue outer
                    }
                    if able{
                        mylog.Println("Farm upgrade started for farm", i, "(probably)...", "Sleeping for 10 minutes")
                        time.Sleep(10*time.Minute)
                        break
                    }
                }
                mylog.Println("Not enought resources available for upgrade... Sleeping for 10 minutes")
                time.Sleep(10*time.Minute)
            } else {
                mylog.Println("Upgrade in progress... Sleeping for 10 minutes")
                time.Sleep(10*time.Minute)
            }
        } else {
            mylog.Println("Not logged in... Logging in in 2 minutes")
            time.Sleep(2*time.Minute)
            mylog.Println("Logging in now")
            tclient.login(settings.Url, settings.Name, settings.Password)
            continue
        }
    }
}