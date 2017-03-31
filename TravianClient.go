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
func (t *TravianClient) login(name, password string){
    //Get main page content
    content:=t.Get_content_and_wait("http://ts4.travian.de")

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
}

var find_costs *regexp.Regexp
var find_upgrade_vars *regexp.Regexp
func (t *TravianClient) try_upgrade(id int) (bool, error){
    //assert id is within good range
    if id<1 || id>38{
        panic("id out of range")
    }

    //Get content of build page
    content:=t.Get_content_and_wait(fmt.Sprintf("http://ts4.travian.de/build.php?id=%d", id))
    //Get current resources
    err:=t.tdata.gather_resource_data(content)
    if err!=nil{
        return false, err
    }

    //Get upgrades cost
    wood, clay, iron, korn, free_korn, err:=func() (int64, int64, int64, int64, int64, error){
        if find_costs==nil{
            re:=""+
            "<span class=\"resources r1.*title=\"Holz\".*alt=\"Holz\" />([0-9]*)</span>"+
            "<span class=\"resources r2.*title=\"Lehm\".*alt=\"Lehm\" />([0-9]*)</span>"+
            "<span class=\"resources r3.*title=\"Eisen\".*alt=\"Eisen\" />([0-9]*)</span>"+
            "<span class=\"resources r4.*title=\"Getreide\".*alt=\"Getreide\" />([0-9]*)</span>"+
            "<span class=\"resources r5.*title=\"freies Getreide\".*alt=\"freies Getreide\" />([0-9]*)</span>"
            find_costs=regexp.MustCompile(re)
        }

        matches:=find_costs.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return 0,0,0,0,0, errors.New("find_costs: len(matches)!=1")
        }

        if len(matches[0])!=6{
            return 0,0,0,0,0, errors.New("find_costs: len(matches[0])!=5")
        }

        if len(matches[0][1])<=0 && len(matches[0][2])<=0 && len(matches[0][3])<=0 && len(matches[0][4])<=0 && len(matches[0][5])<=0{
            return 0,0,0,0,0, errors.New("find_costs: len(matches[0][1])<=0 && len(matches[0][2])<=0 && len(matches[0][3])<=0 && len(matches[0][4])<=0 && len(matches[0][5])<=0")
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

    //Check whether there are enougth resources
    if t.tdata.wood<wood || t.tdata.clay<clay || t.tdata.iron<iron || t.tdata.korn<korn || t.tdata.free_korn<free_korn{
        return false, nil
    }

    //Get needed vars for get request that starts upgrade
    dorf_num, upgrade_id, upgrade_var, err:=func(content []byte) (string,string,string,error) {
        if find_upgrade_vars==nil{
            find_upgrade_vars=regexp.MustCompile("dorf([1-2]).php\\?a=([0-9]{1,10})&amp;c=([0-9a-z]*)'")
        }

        matches:=find_upgrade_vars.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return "","","", errors.New("find_upgrade_vars: len(matches)!=1")
        }

        if len(matches[0])!=4{
            return "","","", errors.New("find_upgrade_vars: len(matches[0])!=4")
        }

        if len(matches[0][1])<=0 && len(matches[0][2])<=0 && len(matches[0][3])<=0{
            return "","","", errors.New("find_upgrade_vars: len(matches[0][1])<=0 && len(matches[0][2])<=0 && len(matches[0][3])<=0")
        }

        return string(matches[0][1]), string(matches[0][2]), string(matches[0][3]), nil
    }(content)
    if err!=nil{
        return true, err
    }

    //Start upgrade
    t.Get_content_and_wait(fmt.Sprintf("http://ts4.travian.de/dorf%s.php?a=%s&c=%s", dorf_num, upgrade_id, upgrade_var))

    return true, nil
}

func (t *TravianClient) gather_data_from_dorf1() error{
    content:=t.Get_content_and_wait("http://ts4.travian.de/dorf1.php")
    err:=t.tdata.gather_non_resource_dorf1_data(content)
    if err!=nil{
        return err
    }
    return t.tdata.gather_resource_data(content)
}
