package main

import "net/http"
import "net/url"
import "io/ioutil"
import "regexp"
import "errors"
import "net/http/cookiejar"
import "math/rand"
import "time"
import "fmt"
import "strconv"

//get and post should not panic: Either return empty or error, or wait and retry

type TravianClient struct{
    http.Client
    tdata TravianData
}

func NewTravianClient() TravianClient {
    cookie_jar, err:=cookiejar.New(nil)
    if err!=nil{
        panic(err.Error())
    }
    return TravianClient{http.Client{Jar: cookie_jar}, TravianData{}}
}

//Adds a "sleep" to the Get
func (t *TravianClient) Get_and_wait(url string) (resp *http.Response, err error){
    resp,err=t.Get(url)
    time.Sleep(time.Duration(rand.Int63n(1500)+500)*time.Millisecond)
    return
}

//Adds a "sleep" to the PostForm
func (t *TravianClient) PostForm_and_wait(url string, data url.Values) (resp *http.Response, err error){
    resp,err=t.PostForm(url, data)
    time.Sleep(time.Duration(rand.Int63n(1500)+500)*time.Millisecond)
    return
}

func (t *TravianClient) Get_content_and_wait(url string) []byte{
    resp,err:=t.Get_and_wait(url)
    if err!=nil{
        panic(err.Error())
    }
    content, err:=ioutil.ReadAll(resp.Body)
    if err!=nil {
        panic(err.Error())
    }
    return content
}

func (t *TravianClient) PostForm_get_content_and_wait(url string, data url.Values) []byte{
    resp,err:=t.PostForm_and_wait(url, data)
    if err!=nil{
        panic(err.Error())
    }
    content, err:=ioutil.ReadAll(resp.Body)
    if err!=nil {
        panic(err.Error())
    }
    return content
}

var find_login *regexp.Regexp
func (t *TravianClient) login(t_url, name, password string){
    //Get main page content
    content:=t.Get_content_and_wait(t_url)

    //Get login var from main page content
    login_var, err:=func(content []byte) (string,error) {
        if find_login==nil{
            find_login=regexp.MustCompile("<input type=\"hidden\" name=\"login\" value=\"([0-9]*)\" />")
        }

        matches:=find_login.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return "", errors.New("len(matches)!=1")
        }

        if len(matches[0])!=2{
            return "", errors.New("len(matches[0])!=2")
        }

        if len(matches[0][1])<=0{
            return "", errors.New("len(matches[0][1])<=0")
        }

        return string(matches[0][1]), nil
    }(content)
    if err!=nil{
        panic(err.Error())
    }

    //Do the actual login
    content=t.PostForm_get_content_and_wait("http://ts4.travian.de/dorf1.php", url.Values{
            "name": {name},
            "password": {password},
            "s1": {"Einloggen"},
            "w": {"1366:768"},
            "login": {login_var},
            "lowRes": {"0"},
            })

    err=t.tdata.gather_data(content)
    if err!=nil{
        panic(err.Error())
    }
}

var find_costs *regexp.Regexp
var find_farm_upgrade_var *regexp.Regexp
func (t *TravianClient) try_upgrade_farm(id int) (bool, error){
    if id<1 || id>18{
        panic("id out of range")
    }
    content:=t.Get_content_and_wait(fmt.Sprintf("http://ts4.travian.de/build.php?id=%d", id))
    err:=t.tdata.gather_data(content)
    if err!=nil{
        return false, err
    }

    wood, clay, iron, korn, free_korn, err:=func() (int64, int64, int64, int64, int64, error){
        if find_costs==nil{
            re:="<div class=\"showCosts centeredText\">"+
            "<span class=\"resources r1\" title=\"Holz\"><img class=\"r1\" src=\"img/x\\.gif\" alt=\"Holz\" />([0-9]*)</span>"+
            "<span class=\"resources r2\" title=\"Lehm\"><img class=\"r2\" src=\"img/x\\.gif\" alt=\"Lehm\" />([0-9]*)</span>"+
            "<span class=\"resources r3\" title=\"Eisen\"><img class=\"r3\" src=\"img/x\\.gif\" alt=\"Eisen\" />([0-9]*)</span>"+
            "<span class=\"resources r4\" title=\"Getreide\"><img class=\"r4\" src=\"img/x\\.gif\" alt=\"Getreide\" />([0-9]*)</span>"+
            "<span class=\"resources r5\" title=\"freies Getreide\"><img class=\"r5\" src=\"img/x\\.gif\" alt=\"freies Getreide\" />([0-9]*)</span>"+
            "<div class=\"clear\"></div></div>"
            find_costs=regexp.MustCompile(re)
        }

        matches:=find_costs.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return 0,0,0,0,0, errors.New("len(matches)!=1")
        }

        if len(matches[0])!=6{
            return 0,0,0,0,0, errors.New("len(matches[0])!=5")
        }

        if len(matches[0][1])<=0{
            return 0,0,0,0,0, errors.New("len(matches[0][1])<=0")
        }

        wood, err:=strconv.ParseInt(string(matches[0][1]), 10, 64)
        if err!=nil{
            return 0,0,0,0,0,err
        }
        clay, err:=strconv.ParseInt(string(matches[0][2]), 10, 64)
        if err!=nil{
            return 0,0,0,0,0,err
        }
        iron, err:=strconv.ParseInt(string(matches[0][3]), 10, 64)
        if err!=nil{
            return 0,0,0,0,0,err
        }
        korn, err:=strconv.ParseInt(string(matches[0][4]), 10, 64)
        if err!=nil{
            return 0,0,0,0,0,err
        }
        free_korn, err:=strconv.ParseInt(string(matches[0][5]), 10, 64)
        if err!=nil{
            return 0,0,0,0,0,err
        }

        return wood, clay, iron, korn, free_korn, nil
    }()
    if err!=nil{
        return false, err
    }

    if t.tdata.wood<wood || t.tdata.clay<clay || t.tdata.iron<iron || t.tdata.korn<korn || t.tdata.free_korn<free_korn{
        return false, nil
    }

    farm_upgrade_var, err:=func(content []byte) (string,error) {
        if find_farm_upgrade_var==nil{
            find_farm_upgrade_var=regexp.MustCompile("dorf1.php\\?a=[0-9]{1,10}&amp;c=([0-9a-z]*)'")
        }

        matches:=find_farm_upgrade_var.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return "", errors.New("len(matches)!=1")
        }

        if len(matches[0])!=2{
            return "", errors.New("len(matches[0])!=2")
        }

        if len(matches[0][1])<=0{
            return "", errors.New("len(matches[0][1])<=0")
        }

        return string(matches[0][1]), nil
    }(content)
    if err!=nil{
        return true, err
    }

    content=t.Get_content_and_wait(fmt.Sprintf("http://ts4.travian.de/dorf1.php?a=%d&c=%s", id, farm_upgrade_var))
    err=t.tdata.gather_data(content)
    if err!=nil{
        return true, err
    }

    return true, nil
}

func (t *TravianClient) gather_data_from_dorf1() error{
    content:=t.Get_content_and_wait("http://ts4.travian.de/dorf1.php")
    return t.tdata.gather_data(content)
}
