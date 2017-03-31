package main

import "log"
import "os"
import "time"
import "math/rand"
// import "fmt"
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
            mylog.Println("Could not gather_data_from_dorf1... Trying again in 2 minutes... Error:", err.Error())
            time.Sleep(2*time.Minute)
            continue
        }

        if !tclient.tdata.is_logged_in{
            mylog.Println("Not logged in... Logging in in 2 minutes")
            time.Sleep(2*time.Minute)
            mylog.Println("Logging in now")
            tclient.login(settings.Url, settings.Name, settings.Password)
            continue
        }

        if tclient.tdata.is_logged_in{
            if !tclient.tdata.is_upgrading{
                for _, fields:=range [][]int{settings.Fields_p1, settings.Fields_p2, settings.Fields_p3}{
                    for _, i:=range rand.Perm(len(fields)){ //int64(1);i<=18;i++{
                        able, err:=tclient.try_upgrade(fields[i])
                        if err!=nil{
                            mylog.Println("Upgrade failed for", fields[i], "... Error:", err.Error())
                            continue outer
                        }
                        if able{
                            mylog.Println("Upgrade started for", fields[i], "(probably)...", "Sleeping for 10 minutes")
                            time.Sleep(10*time.Minute)
                            break
                        }
                    }
                }
                mylog.Println("Not enought resources available for upgrade... Sleeping for 10 minutes")
                time.Sleep(10*time.Minute)
            } else {
                mylog.Println("Upgrade in progress... Sleeping for 10 minutes")
                time.Sleep(10*time.Minute)
            }
        }
    }
}