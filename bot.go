package main

import "log"
import "os"
import "time"
import "math/rand"

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
        tclient.gather_data_from_dorf1()
        if tclient.tdata.is_logged_in{
            if !tclient.tdata.is_upgrading{
                for i:=range rand.Perm(18){ //int64(1);i<=18;i++{
                    i+=1
                    able, err:=tclient.try_upgrade_farm(i)
                    if err!=nil{
                        mylog.Println("Farm upgrade failed for farm", i)
                        continue outer
                    }
                    if able{
                        mylog.Println("Farm upgrade started for farm", i, "...", "Sleeping for 10 minutes")
                        time.Sleep(10*time.Minute)
                        break
                    }
                }
                mylog.Println("Not enought resources available for upgrade... Sleeping for 10 minutes")
                time.Sleep(10*time.Minute)
            } else {
                mylog.Println("Farm upgrade in progress... Sleeping for 10 minutes")
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